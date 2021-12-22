
redo-ifchange package.json package-lock.json

exec >&2

# Don't touch lockfiles; preserve CI compatibility.
npm install --from-lock-file --no-save --include=dev 
touch "$3"