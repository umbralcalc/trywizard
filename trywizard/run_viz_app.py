from base_params import FLOAT_PARAMS, INT_PARAMS
from pyapi.core.spawn_processes import spawn_worldsoop_processes_from_configs
from pyapi.core.implementation_wrappers import (
    OutputCondition,
    OutputFunction,
    TerminationCondition,
    TimestepFunction,
)
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
            OtherParams(float_params=FLOAT_PARAMS, int_params=INT_PARAMS),
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
            iterations=[["rugbyMatch"]],
            output_condition=OutputCondition.every_step(),
            output_function=OutputFunction.stdout(),
            termination_condition=TerminationCondition.number_of_steps(100),
            timestep_function=TimestepFunction.constant(1.0),
        ),
        agent_by_partition={},
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
