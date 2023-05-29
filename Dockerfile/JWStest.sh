#!/bin/bash

password=$2
user=$1
host=$3
path=$4
remoteHost=$5

# 使用参数进行后续操作
echo "Password: $password"
echo "User: $user"
echo "Host: $host"
echo "Path: $path"
echo "RemoteHost: $remoteHost"

# 生成ssh密码
spawn ssh-keygen -t rsa -P '' -f ~/.ssh/id_rsa
expect "Overwrite (y/n)?"  { send "y\n" }

expect <<EOF
  spawn ssh-copy-id $user@$host
  expect {
                "yes/no" { send "yes\n"; exp_continue }
                "password" { send "$password\n"; exp_continue}
                "Password" { send "$password\n" }
          }
  expect eof
EOF

# shellcheck disable=SC2086
sudo scp -r $path $user@$host:$remoteHost