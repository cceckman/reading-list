
export enum State {
  WORKING = "…",
  ERROR = "!",
  OK = "✓",
}

function classOf(s: State) {
  switch (s) {
    case State.WORKING: return "workingState";
    case State.ERROR: return "errorState";
    case State.OK: return "okState";
  };
}


// Status view.
export class Status {
  constructor(container: HTMLElement) {
    this.container = container;

    this.text = document.createTextNode('');
    this.p = document.createElement('p') as HTMLParagraphElement;
    this.p.replaceChildren(this.text);
    this.state = State.WORKING;
    this.update(State.WORKING, "Loading");

    this.container.replaceChildren(this.p);
  }

  update(state: State, message: String) {
    const msg = `${state} ${message}`;
    console.log("Status update: ", msg);
    this.text.data = msg;
    this.container.classList.replace(classOf(this.state), classOf(state));
    this.state = state;
  }

  private state: State;
  private p: HTMLParagraphElement;
  private text: Text;
  private container: HTMLElement;

}