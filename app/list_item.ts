import { Entry } from './entry';

// A view for an item in the reading list.
export class ListItem extends HTMLElement {
  constructor() {
    super();

    const template = document.getElementById('listItem') as HTMLTemplateElement;
    const content = template.content;


    this.attachShadow({ mode: 'open' })
      .appendChild(content.cloneNode(true));
  }

  // Canonical encoding of an Entry (data structure) to a ListItem.
  static create(entry: Entry): ListItem {
    // TODO: Extract this from "create" to "set entry".
    const item = document.createElement("list-item") as ListItem;

    // TODO: This doesn't seem to work; the DOM confirms that the <input> maps into the slot,
    // but it doesn't get included in the form output.
    // We can work around by manually patching the button to include the ID:
    {
      const idSlot = document.createElement("input");
      idSlot.slot = "originalId";
      idSlot.type = "hidden";
      idSlot.name = "originalId";
      idSlot.value = entry.id;
      item.appendChild(idSlot);
    }

    // We can work around this by patching the DOM to include the ID in the button:
    {
      const editButton = item.shadowRoot!.querySelector("button.editButton")! as HTMLButtonElement;
      editButton.value = entry.id;
    }

    // This doesn't work either - having a layer of indirection:
    /*
    {
      const idSlot = document.createElement("span");
      idSlot.slot = "originalId";
      const input = document.createElement("input");
      input.type = "hidden";
      input.name = "originalId";
      input.value = entry.id;

      idSlot.appendChild(input);
      item.appendChild(idSlot);
    }*/
    // Maybe because the <input> is notionally "outside" the form, so it doesn't count?

    {
      const editButton = item.shadowRoot!.querySelector("button.editButton")! as HTMLButtonElement;
      editButton.value = entry.id;
    }


    {
      const titleSlot = document.createElement("span");
      titleSlot.slot = "title";
      titleSlot.replaceChildren(entry.title);
      item.appendChild(titleSlot);
    }

    {
      const sourceSlot = document.createElement("span");
      sourceSlot.slot = "source";
      let sourceText: HTMLElement | string;
      if (entry.source.url) {
        const src = document.createElement("a");
        src.replaceChildren(entry.source.text);
        src.setAttribute("href", entry.source.url.toString());
        src.setAttribute("target", "_blank");
        sourceText = src;
      } else {
        sourceText = entry.source.text;
      }
      sourceSlot.replaceChildren("via ", sourceText);
      item.appendChild(sourceSlot);
    }

    {
      const addedSlot = document.createElement("span");
      addedSlot.slot = "discoveredDate";
      // Trim to just date, not time.
      addedSlot.replaceChildren(entry.added.toISOString().substring(0, 10));
      item.appendChild(addedSlot);
    }

    return item;
  }

  private edit() {
    console.log("editing...");
  }
}

customElements.define('list-item', ListItem);