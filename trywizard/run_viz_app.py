from pyapi.core.spawn_processes import spawn_worldsoop_processes_from_configs
from pyapi.core.config_builder import (
    OtherParams,
    SimulatorImplementationsConfig,
    StochadexImplementationsConfig,
    StochadexSettingsConfig,
    WorldsoopConfig,
    DashboardConfig,
)


def main():
    settings = StochadexSettingsConfig(
        other_params=[
            OtherParams(
                float_params={
                    # Front row possesssion attributes [1 Home, 2 Home, 3 Home, 1 Away, 2 Away, 3 Away]
                    "front_row_scrum_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "front_row_lineout_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "front_row_ruck_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "front_row_maul_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    # Second row possesssion attributes [4 Home, 5 Home, 4 Away, 5 Away]
                    "second_row_scrum_possessions": [1.0, 1.0, 1.0, 1.0],
                    "second_row_lineout_possessions": [1.0, 1.0, 1.0, 1.0],
                    "second_row_ruck_possessions": [1.0, 1.0, 1.0, 1.0],
                    "second_row_maul_possessions": [1.0, 1.0, 1.0, 1.0],
                    # Back row possesssion attributes [6 Home, 7 Home, 8 Home, 6 Away, 7 Away, 8 Away]
                    "back_row_lineout_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "back_row_ruck_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "back_row_maul_possessions": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    # Attributes for all players [1-15 Home, 1-15 Away]
                    "player_run_possessions": 30 * [1.0],
                    "player_attacking_run_scales": 30 * [1.0],
                    "player_defensive_run_scales": 30 * [1.0],
                    "player_fatigue_rates": 30 * [0.0],
                    "player_start_times": 30 * [0.0],
                    # Centres attributes [12 Home, 13 Home, 12 Away, 13 Away]
                    "centres_ruck_possessions": [1.0, 1.0, 1.0, 1.0],
                    "centres_kick_regains": [1.0, 1.0, 1.0, 1.0],
                    # Halves attributes [9 Home, 10 Home, 9 Away, 10 Away]
                    "halves_kick_scales": [1.0, 1.0, 1.0, 1.0],
                    "halves_kick_accuracies": [1.0, 1.0, 1.0, 1.0],
                    # Back three attributes [11 Home, 14 Home, 15 Home, 11 Away, 14 Away, 15 Away]
                    "back_three_kick_scales": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "back_three_kick_accuracies": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "back_three_kick_regains": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    # Lateral run scale is Std of runs 
                    "lateral_run_scale": [1.0],
                    # Probabilities of success for kicking at goal [Home, Away]
                    "goal_probabilities": [0.75, 0.75],
                    # Match event rates
                    "background_event_rates": [1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0],
                    "max_possession_change_rates": [0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2, 0.2],
                    # From Penalty to [Scrum, Kick Phase, Goal, Try]
                    "transition_probs_from_0": [0.25, 0.25, 0.25, 0.25],
                    # From Free Kick to [Scrum, Kick Phase, Run Phase]
                    "transition_probs_from_1": [0.333, 0.333, 0.333],
                    # From Goal to [Kickoff]
                    "transition_probs_from_2": [1.0],
                    # From Drop Goal to [Kickoff]
                    "transition_probs_from_3": [1.0],
                    # From Try to [Kickoff]
                    "transition_probs_from_4": [1.0],
                    # From Kick Phase to [Knock-on, Try, Lineout, Ruck, Maul, Run Phase, Drop Goal]
                    "transition_probs_from_5": [0.142, 0.142, 0.142, 0.142, 0.142, 0.142, 0.142],
                    # From Run Phase to [Knock-on, Try, Lineout, Ruck, Maul]
                    "transition_probs_from_6": [0.2, 0.2, 0.2, 0.2, 0.2],
                    # From Knock-on to [Scrum] 
                    "transition_probs_from_7": [1.0],
                    # From Scrum to [Kick Phase, Run Phase]
                    "transition_probs_from_8": [0.5, 0.5],
                    # From Lineout to [Kick Phase, Run Phase, Ruck, Maul]
                    "transition_probs_from_9": [0.25, 0.25, 0.25, 0.25],
                    # From Ruck to [Kick Phase, Run Phase],
                    "transition_probs_from_10": [0.5, 0.5],
                    # From Maul to [Kick Phase, Run Phase]
                    "transition_probs_from_11": [0.5, 0.5],
                    # From Kickoff to [Kick Phase]
                    "transition_probs_from_12": [1.0],
                },
                int_params={
                    # From Penalty to [Scrum, Kick Phase, Goal, Try]
                    "transitions_from_0": [8, 5, 2, 4],
                    # From Free Kick to [Scrum, Kick Phase, Run Phase]
                    "transitions_from_1": [8, 5, 6],
                    # From Goal to [Kickoff]
                    "transitions_from_2": [12],
                    # From Drop Goal to [Kickoff]
                    "transitions_from_3": [12],
                    # From Try to [Kickoff]
                    "transitions_from_4": [12],
                    # From Kick Phase to [Knock-on, Try, Lineout, Ruck, Maul, Run Phase, Drop Goal]
                    "transitions_from_5": [7, 4, 9, 10, 11, 6, 3],
                    # From Run Phase to [Knock-on, Try, Lineout, Ruck, Maul]
                    "transitions_from_6": [7, 4, 9, 10, 11],
                    # From Knock-on to [Scrum]
                    "transitions_from_7": [8],
                    # From Scrum to [Kick Phase, Run Phase]
                    "transitions_from_8": [5, 6],
                    # From Lineout to [Kick Phase, Run Phase, Ruck, Maul]
                    "transitions_from_9": [5, 6, 10, 11],
                    # From Ruck to [Kick Phase, Run Phase]
                    "transitions_from_10": [5, 6],
                    # From Maul to [Kick Phase, Run Phase]
                    "transitions_from_11": [5, 6],
                    # From Kickoff to [Kick Phase]
                    "transitions_from_12": [5],
                },
            ),
        ],
        init_state_values=[
            [12, 0, 50.0, 35.0, 0, 0, 12, 12, 0, 0, 1.0],
        ],
        init_time_value=0.0,
        seeds=[563],
        state_widths=[11],
        state_history_depths=[2],
        timesteps_history_depth=2,
    )
    implementations = StochadexImplementationsConfig(
        simulator=SimulatorImplementationsConfig(
            iterations=[r"rugbyMatch"],
            output_condition=r"&simulator.EveryStepOutputCondition{}",
            output_function=r"&simulator.StdoutOutputFunction{}",
            termination_condition=(
                r"&simulator.NumberOfStepsTerminationCondition{MaxNumberOfSteps: 100}"
            ),
            timestep_function=r"&simulator.ConstantTimestepFunction{Stepsize: 1.0}",
        ),
        agents=[],
        extra_vars_by_package=[
            {
                "github.com/worldsoop/trywizard/pkg/rugby_simulation": [
                    {"rugbyMatch": r"&rugby_simulation.RugbyMatchIteration{}"},
                ],
            }
        ],
    )
    dashboard = DashboardConfig(
        address=":2112",
        handle="/dashboard",
        millisecond_delay=200,
        react_app_location="viz/",
        launch_dashboard=True,
    )
    config = WorldsoopConfig(
        settings=settings,
        implementations=implementations,
        dashboard=dashboard,
    )
    spawn_worldsoop_processes_from_configs(
        max_workers=1,
        configs=[config],
    )


if __name__ == "__main__":
    main()
