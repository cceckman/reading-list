import { State, Status } from './status';

class App {
  constructor() {
    this.root = document.getElementById("app") as HTMLElement;
    {
      const statusContainer = document.createElement("div") as HTMLDivElement;
      statusContainer.classList.add("statusBar");
      this.status = new Status(statusContainer);
      this.root.replaceChildren(statusContainer);
    }

    if (!('serviceWorker' in navigator)) {
      this.status.update(State.ERROR, "Service worker not available; sharing will not function");
    } else {
      this.status.update(State.WORKING, "Registering service worker");
      navigator.serviceWorker.register('worker.js')
        .then((reg) => {
          this.status.update(State.OK, "Service worker registration succeeded");
        }).catch((error) => {
          this.status.update(State.ERROR, "Service worker registration failed");
          console.log(error)
        });
    }
  }

  private root: HTMLElement;
  private status: Status;
}

new App();