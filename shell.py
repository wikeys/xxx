# -*- coding:UTF-8 -*-
import time
import sys,getopt
import os

try:
	import paramiko
except Exception as e:
	print("请安装paramiko库，pip install paramiko")

class ssh_login:
	def __init__(self,hostname,username,password):
		self.hostname = hostname
		self.username = username
		self.password = password
		ssh = paramiko.SSHClient()
		ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
		ssh.connect(hostname=self.hostname,port=22,username=self.username,password=self.password,timeout=5)
		chan = ssh.invoke_shell()
		self.chan = chan
	def get_message(self):
		lists = [ 
		"dis device",    
		"dis power",
		"dis fan",
		"dis environment",
		"dis cpu",
		"display memory",
		"display logbuffer summary",
		"display users",
		"display mac-address count",
		"display arp all",
		"display ip routing-table",
		"display interface brief",
		"display counters inbound  interface"
		]   #需要执行的命令
		for i in lists:
			if self.chan:
				s = (i + "\n"+"  ")
				self.chan.send(s)
				time.sleep(2)
		res = self.chan.recv(999999)
		print(res.decode())
		self.ssh.close()

if __name__ == '__main__':
	try:
		if len(sys.argv[1:]) == 0:
			print("python "+os.path.basename(__file__)+"  -h help -i 'ip'  -p 'password' -u 'username'")
			sys.exit()	
		opts,args = getopt.getopt(sys.argv[1:],"Hi:u:p:",["help","ip=","passwd=","username="])
	except getopt.GetoptError as err:
		print(os.path.basename(__file__)+"-i 'ip'  -p 'password' -u 'username'")
		sys.exit(2)
	for opt,arg in opts:
		if opt == '-H':
			print("python "+os.path.basename(__file__)+"-i 'ip'  -p 'password' -u 'username'")
			sys.exit()
		elif opt in ("-i","--Hostname"):	
			hostname = arg
		elif opt in ("-u","--username"):
			username = arg
		elif opt in ("-p","--password"):
			password = arg
	run = ssh_login(hostname,username,password)
	run.get_message()
