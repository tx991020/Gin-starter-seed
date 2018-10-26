

# Client Help

<pre>
$ ./client help
Simple client to interact with Dozer API service.

Usage:
  client
  client [command]

Available Commands:
  add         Add api call.
  help        Help about any command
  lookup      Lookup a task uuid.
  mul         Multiply api call.
  token       Print a JWT token.
  version     Print the version.

Use "client [command] --help" for more information about a command.
</pre>

# Client adding
<pre>
$ ./client add --i 1,2,3,4,5,6
Result: 21

</pre>

# Client adding when workers down

<pre>
$ ./client add --i 1,2,3
Defered! task_1db8fd1f-aff0-4db6-9c9a-ada3d20cb006

$ ./client lookup --uuid=task_1db8fd1f-aff0-4db6-9c9a-ada3d20cb006
Status : PENDING

</pre>

# When Workers come back online...

<pre>
./client lookup --uuid=task_1db8fd1f-aff0-4db6-9c9a-ada3d20cb006
Status : SUCCESS
Result : 6
</pre>
