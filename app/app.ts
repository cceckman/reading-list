import { State, Status, StatusBar, workingStatus, errorStatus, okStatus } from './status';

// Event from the 'beforeinstallprompt' event
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<boolean>,
}

class App {
  constructor() {
    this.root = document.getElementById("app") as HTMLElement;
    {
      const statusContainer = document.createElement("div") as HTMLDivElement;
      statusContainer.classList.add("statusBar");
      this.appStatus = new StatusBar(statusContainer);
      this.root.replaceChildren(statusContainer);
    }
    this.appInstalled = okStatus("App installed");
    this.workerInstalled = workingStatus("Checking worker installation");
    this.updateAppStatus();
    this.setupInstallableFlow();
    this.setupServiceWorker();
  }

  private updateAppStatus() {
    if (this.workerInstalled.state != State.OK) {
      this.appStatus.update(this.workerInstalled);
      return;
    }

    this.appStatus.update(this.appInstalled);
  }

  private setupInstallableFlow() {
    // Set up event listeners
    // From https://web.dev/customize-install/:

    // Listen for prompt installation; display an "install me" prompt.
    window.addEventListener('beforeinstallprompt', (e: Event) => {
      console.log("Running install prompt hook");
      e.preventDefault();
      const evt = e as BeforeInstallPromptEvent;

      const install = (() => {
        const install = document.createElement("button") as HTMLButtonElement;
        install.type = "button";
        install.appendChild(document.createTextNode("Install"));
        install.onclick = () => {
          evt.prompt().then((wasInstalled) => {
            if(!wasInstalled) {
              this.appInstalled = errorStatus("Install declined; share-to-list not available");
              this.updateAppStatus();
            }
          });
        };
        return install;
      })();
      const cancel = (() => {
        const b = document.createElement("button") as HTMLButtonElement;
        b.type = "button";
        b.appendChild(document.createTextNode("Ignore"));
        b.onclick = () => {
          this.appInstalled = okStatus("App installation declined");
          this.updateAppStatus();
        };
        return b;
      })();

      const container = document.createElement("span") as HTMLSpanElement;
      container.replaceChildren(
        "Install to share to reading list: ",
        install, cancel,
      );
      this.appInstalled = workingStatus(container);
      this.updateAppStatus();
    });

    // Get notified if it's been installed; set status to OK.
    window.addEventListener('appinstalled', (e: Event) => {
      console.log("Running install-completed hook", e);
      this.appInstalled = okStatus("App installed");
      this.updateAppStatus();
    });
  }

  private setupServiceWorker() {
    if (!('serviceWorker' in navigator)) {
      this.workerInstalled = errorStatus("Service worker not available; sharing will not function");
    } else {
      this.workerInstalled = workingStatus("Registering service worker");
      this.updateAppStatus();
      navigator.serviceWorker.register('worker.js')
        .then((reg) => {
          this.workerInstalled = okStatus("Service worker registered");
          this.updateAppStatus();
        }).catch((error) => {
          console.log(error);
          this.workerInstalled = errorStatus("Service worker registration failed");
          this.updateAppStatus();
        });
    }
  }

  // Does the app need installation?
  private appInstalled: Status;
  // Was there a problem starting the service worker?
  private workerInstalled: Status;

  // TODO:
  // private listView: ListView;
  // private editView: EditView;

  private root: HTMLElement;
  private appStatus: StatusBar;
}

new App();
