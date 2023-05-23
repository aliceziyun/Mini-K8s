#!/bin/bash

# 获取用户输入的密码
# shellcheck disable=SC2162
read -s -p "Enter password: " password
echo

# 获取用户输入的用户名
read -p "Enter username: " username

# 获取用户输入的Host IP
read -p "Enter Host IP: " host_ip

# 输出获取到的信息
echo "Password: $password"
echo "Username: $username"
echo "Host IP: $host_ip"