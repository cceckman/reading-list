

// "Source" type: Where did something come from?
class WebSource {
  constructor(text: String, url?: URL) {
    this.text = text;
    this.url = url;
  }

  text: String;
  url?: URL;
}

// Generalized source: may also include ISBN, DOI, etc.
type Source = WebSource;

class Entry {
  constructor(id: String, title: String, source: Source) {
    this.id = id;
    this.title = title;
    this.source = source;
    this.added = new Date();
  }

  // Unique identifier for this entry.
  id: String;
  // Renderable text: What do we call it?
  title: String;
  // How can it be found?
  source: Source;
  // When was this entry added to the list?
  added: Date;

  // Who wrote this entry? Where can they be found on the Internet?
  author?: Source;

  // How did I find out about this item?
  discovery?: Source;

  // When was this entry marked "read"?
  read?: Date;
  // When was this entry marked "reviewed", i.e. commentary ready for sharing?
  reviewed?: Date;
}