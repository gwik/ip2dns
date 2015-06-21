dyniptoroute53 is a command line program to set your public ip a DNS record in AWS Route 53.

*The record must be created first.*

Usage:
```sh
	# dyniptoroute53 -help
	Usage of dyniptoroute53:
	  -accesskey="": AWS access key
	  -host="": [REQUIRED] Host name to set.
	  -lock="/tmp": Path for the lock file.
	  -secretkey="": AWS secret key
	  -zoneid="": [REQUIRED] AWS Route53 zone ID.
```

Example:
```sh
	dyniptoroute53 -host yourdomain.tld -zoneid "..." -accesskey "..." -secretkey "..." 2> /tmp/dyniptoroute53.log
```

Example crontab :
```sh
*/5 * * * * /usr/local/bin/dyniptoroute53 -host yourdomain.tld -zoneid "..." -accesskey "..." -secretkey "..." 2> /tmp/dyniptoroute53.log
```
