
redo-ifchange package.json package-lock.json

exec >&2

npm install --include=dev 
touch "$3"