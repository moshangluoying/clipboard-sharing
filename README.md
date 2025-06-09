# shear-plate-sharing
在原作者的基础上使用AI修复了bug：复制文本内容过多（实测超800行代码文本）就会崩溃；

## What is that?
when you use different os in different computer,you may need share shear plate,
so this program is the answer.

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

### 2、build and run
#### Windows
Build options:
```powershell
# Build with console window (default)
.\build.ps1

# Build without console window
.\build.ps1 -NoConsole

# Build release version with console window
.\build.ps1 -Release

# Build release version without console window
.\build.ps1 -Release -NoConsole
```

Run the program as administrator (required for keyboard state synchronization):
```powershell
# Right-click plate.exe and select "Run as administrator"
```

#### Linux
```shell
./plate
```

Note: On Linux systems, you need to install xdotool for keyboard state synchronization:
```shell
# Debian/Ubuntu
sudo apt-get install xdotool

# Fedora
sudo dnf install xdotool

# Arch Linux
sudo pacman -S xdotool
```

## todo
1、make the code perfect  
2、support copy img

## todo-lanxin
1、sync num_lock and capslock
