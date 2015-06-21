

**dyniptoroute53** is a simple command line tool to set your public IP address on a AWS route 53 record.

**The record must be created first.**

Usage:
```sh
dyniptoroute53 -help
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
```crontab
*/5 * * * * /usr/local/bin/dyniptoroute53 -host yourdomain.tld -zoneid "..." -accesskey "..." -secretkey "..." 2> /tmp/dyniptoroute53.log
```
