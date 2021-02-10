# Hackernews Clone GraphQL API

This is the GraphQL API for a hackernews clone. This API is written in Go and uses the
graph-gophers set of packages to import the schema and write resolvers for them. This was done
as a learning project for teaching myself GraphQL. The web fronted for this, using 
VueJS and Apollo, can be found [here](https://github.com/leggettc18/hackernews-vue-apollo).

## Building

### Prerequisites
1. Go (obviously)
2. GCC (the sqlite package I used utilizes cgo, so it needs a C compiler)

### Building
If you have the prerequisites met just run `go build -i` with whatever other arguments you wish.
(i.e. -o for naming the exe, or -ldflags needed for debugging, etc.) If you need any -ldflags
for debugging, you may need to run `go build -i` once first without it to compile all the dependencies
without it, then omit the `-i` afterwards. Not sure if that's necessary, I know I've run into that
with another Go project in the past.

## Running
After building steps above, just run the executable generated in the same directory (or whatever
directory you specified in the `-o` argument to `go build`). The sqlite database will be
initialized each time the executable is run. The default login is `admin@example.com` with
the password `password`. 

## Feedback
Bear in mind this was done as an exercise for learning GraphQL. Code quality may not be perfect
and there will probably be bugs. That being said, in the interest of improving and being a better
resource for others, any feedback to any aspect of this is appreciated. Whether that be the code itself
or some missed prerequisites or improvements to the build process, etc. If you have any feedback, leave
it  as an issue on this repo.
