
class Installer {
  constructor() {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('worker.js')
        .then((reg) => {
          console.log('Registration succeeded; got: ', reg.scope);
        }).catch((error) => {
          console.log('Registration failed: ', error);
        });
    }    
  }
}

new Installer();