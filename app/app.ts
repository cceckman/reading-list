import { State, Status, StatusBar, workingStatus, errorStatus, okStatus } from './status';

// Event from the 'beforeinstallprompt' event
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<boolean>,
}

// Extensions to the navigator:
// https://wicg.github.io/get-installed-related-apps/spec/
interface ExtendedNavigator extends Navigator {
  getInstalledRelatedApps: () => Promise<Array<{ id: String }>>;
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

    this.appInstalled = workingStatus("Checking app installation");
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
    const installSuccess = () => {
      const appInstalledStatus = okStatus("App installed");
      console.log("App appears installed");
      this.appInstalled = appInstalledStatus;
      this.appInstallPrompt = undefined;
      this.updateAppStatus();
    };

    // If we're already standalone, assume we've been installed.
    // This isn't actually correct, of course.
    if(window.matchMedia('(display-mode: standalone)').matches){
      console.log("Assuming standalone window indicates installation");
      installSuccess();
      return;
    }

    const installPrompt = (evt?: BeforeInstallPromptEvent) => {
        // Not already installed; capture the prompt.
        // TODO: make the message a node, to properly prompt
        this.appInstalled = workingStatus("Click to install app");
        this.appInstallPrompt = evt;
        this.updateAppStatus();
    };

    // Set up event listeners
    // From https://web.dev/customize-install/:
    // Get a handle if it needs to be installed.
    window.addEventListener('beforeinstallprompt', (e: Event) => {
      console.log("Running install prompt hook");
      e.preventDefault();
      if (this.appInstalled.state != State.OK) {
        installPrompt(e as BeforeInstallPromptEvent);
      }
    });
    // Get notified if it's been installed; set status to OK.
    window.addEventListener('appinstalled', (e: Event) => {
      console.log("Running app install hook", e);
      installSuccess();
    });

    // If we can check-for-installation, go ahead and do so-
    // and short-circuit events if we find ourself.
    if ('getInstalledRelatedApps' in window.navigator) {
      const nav = navigator as ExtendedNavigator;
      nav.getInstalledRelatedApps().then((apps) => {
        console.log("Installed apps: ", apps);
        if (apps.length != 0) {
          installSuccess();
        }
      });
    } else {
      console.log("getInstalledRelatedApps not available");
    }
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

  // Has the app been installed, i.e. as a share target?
  private appInstalled: Status;
  private appInstallPrompt?: BeforeInstallPromptEvent;
  private workerInstalled: Status;

  private root: HTMLElement;
  private appStatus: StatusBar;
}

new App();
