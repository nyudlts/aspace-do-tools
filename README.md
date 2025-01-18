## do-tools

This repo contains command-line-interface (CLI) code for interacting    
with elements of an [ArchivesSpace](https://github.com/archivesspace/archivesspace) instance. 


#### Running
```
$ do-tools -h     # prints out available commands

# example commands 
$ do-tools do refresh --ao-uri /repositories/11/archival_objects/23745 -c /path/to/config.yaml -e test 
$ do-tools do update   -a      /repositories/3/archival_objects/512700 -o "https://hdl.handle.net/2333.1/material-request-placeholder" -f "https://hdl.handle.net/2333.1/djh9w43b" -u "video-service"  -c cmd/testsupport/config.yaml -e test
```

**NOTE:**   
When using **Windows** Bash shells, e.g., Git Bash, you will want to   
`export MSYS_NO_PATHCONV=1` to prevent the shell from trying to expand     
an archival object URI into a filesystem path. 

For example,  
if `MSYS_NO_PATHCONV=1` is not set, the `ao-uri` argument is changed as follows:    
```
                    /repositories/11/archival_objects/23745 
--> 
C:/Program Files/Git/repositories/11/archival_objects/23745  
```


#### Testing
The current test suite assumes that you have a file named `config.yaml`   
in the `cmd/testsupport` directory with an environment named `test`.    

There is a template file in the `cmd/testsupport` directory that can    
be used to set up `config.yaml` for your specific test environment.     

Note that `cmd/testsupport/config.yaml` is **ignored** by Git via the    
`.gitignore` file to avoid accidentally committing application secrets.    

Once the `config.yaml` file is set up, you can `cd` into `cmd/` and run    
`$ go test -v`

---