
# PD Project

在bash中，$( )与` `（反引号）都是用来作命令替换的。$( )的弊端是，并不是所有的类unix系统都支持这种方式，但反引号是肯定支持的。
## set cmd
set -e表示一旦脚本中有命令的返回值为非0，则脚本立即退出，后续命令不再执行;
set -o pipefail表示在管道连接的命令序列中，只要有任何一个命令返回非0值，则整个管道返回非0值，即使最后一个命令返回0.