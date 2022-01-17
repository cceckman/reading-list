import { Entry } from "./entry";

export class EntryView {
    constructor(mainQuery: string) {
        const mainTemplate : HTMLTemplateElement = document.querySelector(mainQuery) as HTMLTemplateElement;

        this.root = mainTemplate.content.cloneNode(true) as HTMLElement;
    }

    display(e: Entry) {
        // TODO: Fill template
    }

    root : HTMLElement;
}