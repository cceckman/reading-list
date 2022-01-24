
// A view for an item in the reading list.
class ListItem extends HTMLElement {
  constructor() {
    super();

    const template = document.getElementById('listItem') as HTMLTemplateElement;
    const content = template.content;

    this.attachShadow({ mode: 'open' })
      .appendChild(content.cloneNode(true));
  }

  private edit() {
    console.log("editing...");
  }
}

export function AddListItem() {
  customElements.define('list-item', ListItem);
}