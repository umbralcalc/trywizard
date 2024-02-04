<p align="center">
<img src="../viz/src/assets/trywizard-logo-animated.gif" width="400"/>
</p>

Trywizard a simple event-based rugby match management simulator written in `Go` (see the [stochadex](https://github.com/umbralcalc/stochadex) package) which accompanies a 2D pitch visualisation. For more details on how the match engine was created, you can read [this chapter](https://umbralcalc.github.io/worlds-of-observation/managing_a_rugby_match/chapter.pdf) in the open source book: [Worlds of Observation](https://umbralcalc.github.io/worlds-of-observation/).

<p align="center">
<img src="../viz/src/assets/pitch-background.png" width="400"/>
</p>

## How to setup

Get the python environment sorted

```bash
python3 -m venv venv
source venv/bin/activate
pip3 install -r requirements.txt
export PYTHONPATH={$PYTHONPATH}:your/path/to/worldsoop
export WORLDSOOP_PATH=your/path/to/worldsoop
```

## How to run the visualisation app

```shell
# install the dependencies of and build the app
cd ./viz && npm install && npm run build && cd ..

# run the main script and checkout http://localhost:3000
python trywizard/run_viz_app.py
```
