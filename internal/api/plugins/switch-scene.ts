import OBSWebSocket from "obs-websocket-bun";
const obs = new OBSWebSocket();

const BUNDECK_DEVICES = ["~~Webcam", "~~WebcamCS"];
const BUNDECK_OBS_PASSWORD = "password";
const BUNDECK_OBS_PORT = 4455;

async function toggleWebcam() {
  try {
    await obs.connect(
      `ws://localhost:${BUNDECK_OBS_PORT}`,
      BUNDECK_OBS_PASSWORD
    );

    const scenes = (await obs.call("GetSceneList"))
      .scenes as unknown as Scene[];
    const activeScene = (await obs.call(
      "GetCurrentProgramScene"
    )) as unknown as Scene;

    // Reorder scenes to put active scene first
    const orderedScenes = [
      ...scenes.filter((s) => s.sceneName === activeScene.sceneName),
      ...scenes.filter((s) => s.sceneName !== activeScene.sceneName),
    ];

    // Get active scene items and determine initial state
    const activeSceneItems = (
      await obs.call("GetSceneItemList", {
        sceneName: activeScene.sceneName,
      })
    ).sceneItems as unknown as SceneItem[];

    // Default to true if none of the devices are found in active scene
    let isEnabled = true;

    // Check if any of our devices exist in active scene and get their state
    for (const device of BUNDECK_DEVICES) {
      const activeItem = activeSceneItems.find((i) => i.sourceName === device);
      if (activeItem) {
        isEnabled = activeItem.sceneItemEnabled;
        break; // Use the first found device's state
      }
    }

    let toggled = false;

    // Toggle all devices in all scenes
    for (const scene of orderedScenes) {
      const sceneItems = (
        await obs.call("GetSceneItemList", {
          sceneName: scene.sceneName as string,
        })
      ).sceneItems as unknown as SceneItem[];

      for (const device of BUNDECK_DEVICES) {
        const item = sceneItems.find((i) => i.sourceName === device);

        if (item) {
          await obs.call("SetSceneItemEnabled", {
            sceneName: scene.sceneName as string,
            sceneItemId: item.sceneItemId as number,
            sceneItemEnabled: !isEnabled as boolean,
          });

          toggled = true;
        }
      }
    }

    await obs.disconnect();

    if (toggled) {
      console.log("toggled", !isEnabled ? "on" : "off");
    } else {
      console.log("couldn't find any matching sources");
    }
  } catch (e) {
    console.error((e as Error).message);
  }
}

toggleWebcam();

interface SceneItem {
  inputKind: string | null;
  isGroup: boolean;
  sceneItemBlendMode: string;
  sceneItemEnabled: boolean;
  sceneItemId: number;
  sceneItemIndex: number;
  sceneItemLocked: boolean;
  sceneItemTransform: {
    alignment: number;
    boundsAlignment: number;
    boundsHeight: number;
    boundsType: string;
    boundsWidth: number;
    cropBottom: number;
    cropLeft: number;
    cropRight: number;
    cropToBounds: boolean;
    cropTop: number;
    height: number;
    positionX: number;
    positionY: number;
    rotation: number;
    scaleX: number;
    scaleY: number;
    sourceHeight: number;
    sourceWidth: number;
    width: number;
  };
  sourceName: string;
  sourceType: string;
  sourceUuid: string;
}

interface Scene {
  currentProgramSceneName: string;
  currentProgramSceneUuid: string;
  sceneName: string;
  sceneUuid: string;
}
