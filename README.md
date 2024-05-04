<p align="center">
<img src="../viz/src/assets/trywizard-logo-animated.gif" width="400"/>
</p>

Trywizard a simple event-based rugby match management simulator written with the [stochadex API](https://github.com/umbralcalc/stochadex) which accompanies a 2D pitch visualisation written with React. For more details on how the match engine was created, you can read [this article](https://umbralcalc.github.io/posts/trywizard.html).

<p align="center">
<img src="../viz/src/assets/pitch-background.png" width="400"/>
</p>

## How to setup

```shell
# install the dependencies of and build the app
cd ./viz && npm install && npm run build && cd ..

# run the stochadex binary passing in the configs 
# and checkout http://localhost:3000
/path/to/repo/stochadex/bin/stochadex --config ./cfg/config.yaml \
--dashboard ./cfg/dashboard_config.yaml 
```
