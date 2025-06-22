package rugby_simulation

// func TestRugbyMatch(t *testing.T) {
// 	t.Run(
// 		"test that the Rugby match runs",
// 		func(t *testing.T) {
// 			settings := simulator.LoadSettingsFromYaml("rugby_match_settings.yaml")
// 			iteration := &RugbyMatchIteration{}
// 			iteration.Configure(0, settings)
// 			partitions := []simulator.Partition{{Iteration: iteration}}
// 			implementations := &simulator.Implementations{
// 				Partitions:      partitions,
// 				OutputCondition: &simulator.NilOutputCondition{},
// 				OutputFunction:  &simulator.NilOutputFunction{},
// 				TerminationCondition: &simulator.NumberOfStepsTerminationCondition{
// 					MaxNumberOfSteps: 100,
// 				},
// 				TimestepFunction: &simulator.ConstantTimestepFunction{Stepsize: 1.0},
// 			}
// 			coordinator := simulator.NewPartitionCoordinator(
// 				settings,
// 				implementations,
// 			)
// 			coordinator.Run()
// 		},
// 	)
// }
