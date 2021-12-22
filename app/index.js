"use strict";
(() => {
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __commonJS = (cb, mod) => function __require() {
    return mod || (0, cb[__getOwnPropNames(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };

  // js/status.js
  var require_status = __commonJS({
    "js/status.js"(exports) {
      "use strict";
      Object.defineProperty(exports, "__esModule", { value: true });
      exports.Status = exports.State = void 0;
      var State;
      (function(State2) {
        State2["WORKING"] = "\u2026";
        State2["ERROR"] = "!";
        State2["OK"] = "\u2713";
      })(State = exports.State || (exports.State = {}));
      function classOf(s) {
        switch (s) {
          case State.WORKING:
            return "workingState";
          case State.ERROR:
            return "errorState";
          case State.OK:
            return "okState";
        }
        ;
      }
      var Status = class {
        constructor(container) {
          this.container = container;
          this.text = document.createTextNode("");
          this.p = document.createElement("p");
          this.p.replaceChildren(this.text);
          this.state = State.WORKING;
          this.update(State.WORKING, "Loading");
        }
        update(state, message) {
          const msg = `${state} ${message}`;
          console.log("Status update: ", msg);
          this.text.data = msg;
          this.p.classList.replace(classOf(this.state), classOf(state));
          this.state = state;
        }
      };
      exports.Status = Status;
    }
  });

  // js/app.js
  var require_app = __commonJS({
    "js/app.js"(exports) {
      Object.defineProperty(exports, "__esModule", { value: true });
      var status_1 = require_status();
      var App = class {
        constructor() {
          this.root = document.getElementById("app");
          {
            const statusContainer = document.createElement("div");
            statusContainer.classList.add("statusBar");
            this.status = new status_1.Status(statusContainer);
            this.root.replaceChildren(statusContainer);
          }
          if (!("serviceWorker" in navigator)) {
            this.status.update(status_1.State.ERROR, "Service worker not available; sharing will not function");
          } else {
            this.status.update(status_1.State.WORKING, "Registering service worker");
            navigator.serviceWorker.register("worker.js").then((reg) => {
              this.status.update(status_1.State.OK, "Service worker registration succeeded");
            }).catch((error) => {
              this.status.update(status_1.State.ERROR, "Service worker registration failed");
              console.log(error);
            });
          }
        }
      };
      new App();
    }
  });
  "use strict";
  require_app();
})();
