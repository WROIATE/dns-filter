# DNS filter
A tool for blocking AAAA dns record when domain have both AAAA and A records

```shell
Usage of dns-filter:    
  -l string  
        listen address (default "0.0.0.0:5367")  
  -s string  
        upstream dns server (default "1.1.1.1:53")
```  

## How to use
Just run and send dns query to your listen address:
```shell
dns-filter -s "1.1.1.1:53" -l "0.0.0.0:5367" 
```
Try to verify it:
```shell
$ nslookup -port=5367 www.google.com 127.0.0.1
Server:		127.0.0.1
Address:	127.0.0.1#5367

Non-authoritative answer:
Name:	www.google.com
Address: 142.250.76.132
```
