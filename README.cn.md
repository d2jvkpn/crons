### Crons
---
Crons 程序用于控制一组命令行的定时启动或重启, 支持 windows, linux 平台。

#### 1. Crons
##### 1.1 安装与编译
```bash
git clone git@github.com:d2jvkpn/crons.git
cd crons

make build
ls target/crons*
```

##### 1.2 配置文件 (yaml)
**默认配置文件路径 configs/local.yaml, 可以使用 --config 进行指定**

```yaml
jobs:
# required: name, path, cron
- name: SLEEP
  path: sleep
  args: [70]
  working_dir: ""
  # https://crontab.guru/#30_00_*_*_1
  cron:
    minute: "30"
    hour: "00"
    month_day: "*"
    month: "*"
    week_day: "1"
  max_retries: 15
  start_immediately: true
```

jobs 下一个任务的配置
- *name: string, 任务名称;
- *path: string, 执行文件名称或路径;
- args: []string([]string{}), 命令参数列表;
- working_dir: string(""), 工作目录;
- cron: 定时配置 (https://crontab.guru/#00_00_*_*_1)
  - minute: string("*"), 分钟;
  - hour: string("*"), 小时;
  - month_day: string("*"), 日期;
  - month: string("*"), 月份;
  - week_day: string("*"), 星期 (0 表示周日);
- max_retries: uint(0), 一个定时执行周期内, 运行失败重试最大次数;
- start_immediately: bool(false), 是否直接启动 (即不等待下一次 cron 时间);

*示例中的 cron 配置为每周一 00:30 自动重启*

##### 1.3 运行
```bash
./target/crons -config=configs/local.yaml
```

```powershell
.\crons.exe --config=configs/local.yaml
```

#### 1.4 注意事项
- 运行时, 保持终端窗口打开, 关闭窗口将关闭 crons 程序以及它控制的子进程;

#### 2. web
待开发...
