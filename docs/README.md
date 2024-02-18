<p align="center">
<img src="../viz/src/assets/trywizard-logo-animated.gif" width="400"/>
</p>

Trywizard a simple event-based rugby match management simulator written with the WorldsOOp Go/Python API which accompanies a 2D pitch visualisation written with React and TypeScript. For more details on how the match engine was created, you can read [this article](https://umbralcalc.github.io/posts/trywizard.html).

<p align="center">
<img src="../viz/src/assets/pitch-background.png" width="400"/>
</p>

## How to setup

Get the python environment sorted

```bash
python3 -m venv venv
source venv/bin/activate
pip3 install -r requirements.txt
export WORLDSOOP_PATH=/your/path/to/worldsoop
export PYTHONPATH=${PYTHONPATH}:${WORLDSOOP_PATH}
```

## How to run the visualisation app

```shell
# install the dependencies of and build the app
cd ./viz && npm install && npm run build && cd ..

# run the main script and checkout http://localhost:3000
python trywizard/run_viz_app.py
```
