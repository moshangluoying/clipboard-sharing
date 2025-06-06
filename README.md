# shear-plate-sharing

## What is that?
when you use different os in different computer,you may need share shear plate,
so this program is the answer.

## Requirements

### For Linux Users
You need to install X11 development libraries:
```bash
# Ubuntu/Debian
sudo apt-get install -y libx11-dev

# CentOS/RHEL
sudo yum install -y libX11-devel
```

It's recommended to build the program directly on Linux:
```bash
go build -o plate
```

## How to use?
### 1、edit `config.yml`
server:
```yaml
Role: server
Port: 7777
Password: xxx
```
client:
```yaml
Host: 192.168.31.174
Port: 7777
Password: xxx
```
tips: do not run client when you pc has been run server

### 2、run
```shell
./plate
```

## todo
1、make the code perfect  
2、support copy img

## todo-lanxin
1、sync num_lock and capslock
