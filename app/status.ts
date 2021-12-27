
export function workingStatus(message: String | HTMLElement): Status {
  return {
    state: State.WORKING,
    message,
  }
}

export function errorStatus(message: String | HTMLElement): Status {
  return {
    state: State.ERROR,
    message,
  }
}

export function okStatus(message: String | HTMLElement): Status {
  return {
    state: State.OK,
    message,
  }
}

export interface Status {
  state: State,
  message: String | HTMLElement,
}

export enum State {
  WORKING,
  ERROR,
  OK,
}

function classOf(s: State) {
  switch (s) {
    case State.WORKING: return "workingState";
    case State.ERROR: return "errorState";
    case State.OK: return "okState";
  };
}

function stringOf(s: State) {
  switch(s) {
    case State.WORKING: return "…";
    case State.ERROR: return "!";
    case State.OK: return "✓";
  }
}

// Status view.
export class StatusBar {
  constructor(container: HTMLElement) {
    this.container = container;

    this.p = document.createElement('p') as HTMLParagraphElement;
    this.state = State.WORKING;
    this.update(workingStatus("Loading"));

    this.container.replaceChildren(this.p);
  }

  update({state, message}: Status) {
    if(message instanceof HTMLElement) {
      const text = document.createTextNode(`${stringOf(state)} `);
      this.p.replaceChildren(text, message);
    } else {
      const text = document.createTextNode(`${stringOf(state)} ${message}`);
      this.p.replaceChildren(text);
    }

    this.container.classList.remove(classOf(this.state))
    this.container.classList.add(classOf(state));
    this.state = state;
  }

  private state: State;
  private p: HTMLParagraphElement;
  private container: HTMLElement;

}