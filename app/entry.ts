

// "Source" type: Where did something come from?
export interface WebSource {
  text: string;
  url?: URL;
}

// Generalized source: may also include ISBN, DOI, etc.
export type Source = WebSource;

export interface Entry {
  // Unique identifier for this entry.
  id: string;
  // Renderable text: What do we call it?
  title: string;
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