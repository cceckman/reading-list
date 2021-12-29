import { ListItem } from "./list_item";
import { Entry } from "./entry";

export class ListView {
  constructor(mainQuery: string) {
    const mainTemplate: HTMLTemplateElement = document.querySelector(mainQuery) as HTMLTemplateElement;

    this.root = mainTemplate.content.cloneNode(true) as HTMLElement;
    this.items = this.root.querySelector("#listItems")!;

    {
      const loading: HTMLParagraphElement = document.createElement("p");
      loading.appendChild(
        document.createTextNode("Loading...")
      );
      this.loading = loading;
    }
    this.items.replaceChildren(this.loading);

    {
      const refreshButton: HTMLButtonElement = this.root.querySelector(".menuButton#refresh")!;
      refreshButton.onclick = () => {
        this.refresh()
      };
    }

    {
      const addButton: HTMLButtonElement = this.root.querySelector(".menuButton#add")!;
      addButton.onclick = () => {
        // TODO: actually implement
        this.mockAdd();
      };
    }

    // TODO: refresh right away? Or on-display?
    // this.refresh();
  }

  // TODO: Implement sort / filter.

  // TODO: Implement a real "add", not a mock.
  private mockAdd() {
    if (this.loading.parentElement === this.items) {
      this.items.removeChild(this.loading);
    }
    const n = this.items.childElementCount;
    const entry: Entry = {
      id: `added item ${n}`,
      title: `Manually-added item ${n}`,
      source: {
        text: "Javascript"
      },
      added: new Date(),
    };

    if (n % 2 == 0) {
      entry.source.url = new URL("https://github.com/cceckman/reading-list");
    }

    this.items.appendChild(ListItem.create(entry));
  }

  private refresh() {
    this.items.replaceChildren(this.loading);

    // TODO: actually load. In the mean time...
    if (this.fetcher !== undefined) {
      clearInterval(this.fetcher);
    }
    this.fetcher = setInterval(() => {
      this.mockAdd();
    }, 1000);
  }


  root: HTMLElement;

  private loading: HTMLElement;
  private items: HTMLDivElement;

  // TODO: Remove; this is a stub for mockup.
  private fetcher?: number;
}