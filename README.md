# quest

A curl-like utility for making HTTP requests at the command line

# installation

Clone the repo `git clone https://github.com/matthewdavidrodgers/quest` and navigate to the installed directory
Install `go install ./cmd/quest` (remember to add your go binary location to your path - alternatively, you can also use `go build ./cmd/quest` if you want control of the built binary)

# usage

`quest <options> <url>`

with options as follows:
- `-f` format the output as json (defaults to true)
- `-e` open vim with a temporary file containing the request details before sending, allowing you to fine tune your request. Once you save and close vim, the updated details will be used (defaults to false)
- `-m` specify an HTTP method (allowed: GET, PUT, POST, PATCH, DELETE, defaults to GET)

quest also supports saving and reusing cookies using the following command:

`quest cookie add <my-cookie>`

your cookies are saved to ~/.questconfig and re-used on every request (note that there is not host matching for which domains to send which cookies to based on your passed url)

to reset your cookies:

`quest cookie wipe`
