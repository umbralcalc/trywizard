package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/umbralcalc/stochadex/pkg/interactions"
	"github.com/umbralcalc/stochadex/pkg/simulator"

	"github.com/worldsoop/trywizard/go/rugby_simulation"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startVizApp() (*os.Process, error) {
	cmd := exec.Command("serve", "-s", "build")
	cmd.Dir = "../viz/"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dashboard app: %w", err)
	}

	return cmd.Process, nil
}

type StepperOrRunner interface {
	Run()
	Step(wg *sync.WaitGroup)
	ReadyToTerminate() bool
}

func LoadStepperOrRunner(
	settings *simulator.LoadSettingsConfig,
	implementations *simulator.LoadImplementationsConfig,
	agents []*interactions.AgentConfig,
) StepperOrRunner {
	if len(agents) == 0 {
		return simulator.NewPartitionCoordinator(
			simulator.NewStochadexConfig(
				settings,
				implementations,
			),
		)
	} else {
		return interactions.NewPartitionCoordinatorWithAgents(
			&interactions.LoadConfigWithAgents{
				Settings:        settings,
				Implementations: implementations,
				Agents:          agents,
			},
		)
	}
}

func main() {
	settings := simulator.NewLoadSettingsConfigFromYaml("./rugby_simulation/rugby_match_config.yaml")
	iterations := []simulator.Iteration{&rugby_simulation.RugbyMatchIteration{}}
	for partitionIndex := range settings.StateWidths {
		iterations[partitionIndex].Configure(partitionIndex, settings)
	}
	implementations := &simulator.LoadImplementationsConfig{
		Iterations:           iterations,
		OutputCondition:      &simulator.EveryStepOutputCondition{},
		OutputFunction:       &simulator.NilOutputFunction{},
		TerminationCondition: &simulator.NumberOfStepsTerminationCondition{MaxNumberOfSteps: 100},
		TimestepFunction:     &simulator.ConstantTimestepFunction{Stepsize: 1.0},
	}
	agents := []*interactions.AgentConfig{}
	// agents := []*interactions.AgentConfig{
	// 	{
	// 		Actor:       &interactions.DoNothingActor{},
	// 		Generator:   &interactions.DoNothingActionGenerator{},
	// 		Observation: &interactions.GaussianNoiseStateObservation{},
	// 	},
	// }
	var vizProcess *os.Process
	vizProcess, err := startVizApp()
	if err != nil {
		log.Fatal(err)
	}
	defer vizProcess.Signal(os.Interrupt)
	http.HandleFunc(
		"/dashboard",
		func(w http.ResponseWriter, r *http.Request) {
			connection, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("Error upgrading to WebSocket:", err)
				return
			}
			defer connection.Close()

			var mutex sync.Mutex
			implementations.OutputFunction =
				simulator.NewWebsocketOutputFunction(connection, &mutex)
			stepperOrRunner := LoadStepperOrRunner(settings, implementations, agents)

			var wg sync.WaitGroup
			// terminate the for loop if the condition has been met
			for !stepperOrRunner.ReadyToTerminate() {
				stepperOrRunner.Step(&wg)
				time.Sleep(200 * time.Millisecond)
			}
		},
	)
	log.Fatal(http.ListenAndServe(":2112", nil))
	vizProcess.Signal(os.Interrupt)
	vizProcess.Wait()
}
