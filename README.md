gsupervisor
===========

A guardian for the specified process, it can be restart it when it die 


used like :

gsupervisor <command> args1 args2.....


the command's STDOUT will redirect to outfile defined by supervisor.conf.

gsupervisor like nohup, ignored SIGHUP signal, so it's safe when terminal exit
