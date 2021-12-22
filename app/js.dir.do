
redo-ifchange modules.dir *.ts tsconfig.json

exec >&2

npm run tsc

touch "$3"