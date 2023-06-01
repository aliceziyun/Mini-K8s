import paramiko
import time
import os
from scp import SCPClient

def GPUjob(username,password,host,path,remote):
    #上传到远端
    ssh=paramiko.SSHClient()
    ssh.load_system_host_keys()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(hostname=host, username=username, password=password)
    scpclient = SCPClient(ssh.get_transport(),socket_timeout=100.0)
    local_path = []
    for root,folder_names, file_names in os.walk(path):
        for file_name in file_names:
            local_path.append(path+"/"+file_name)
    remote_path = remote
    scpclient.put(local_path, remote_path)

    #运行slurm脚本
    sshin,sshout,ssherr = ssh.exec_command("sbatch test.slurm")
    res = sshout.read()
    jobID = res.split()[-1].decode('utf-8')


    #每隔一段时间轮询状态
    while True:
        cmd = "sacct -j " + jobID + " | tail -n +3 | awk '{print $1, $2, $3, $4, $5, $6, $7}'"
        sshin,sshout,ssherr = ssh.exec_command(cmd)
        res = sshout.read().split()
        if len(res) == 0:
            continue
        print(res)
        status = res[5].decode('utf-8')
        if status == "COMPLETED":
            scpclient.get(jobID+".out",path)
            scpclient.get(jobID+".err",path)
            break
        else:
            time.sleep(10)