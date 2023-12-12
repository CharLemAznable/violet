### violet

[![Build](https://github.com/CharLemAznable/violet/actions/workflows/go.yml/badge.svg)](https://github.com/CharLemAznable/violet/actions/workflows/go.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/CharLemAznable/violet)

[![MIT Licence](https://badges.frapsoft.com/os/mit/mit.svg?v=103)](https://opensource.org/licenses/mit-license.php)
![GitHub code size](https://img.shields.io/github/languages/code-size/CharLemAznable/violet)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=alert_status)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)

[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=bugs)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)

[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=security_rating)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)

[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=sqale_index)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=code_smells)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)

[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=ncloc)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=coverage)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=CharLemAznable_violet&metric=duplicated_lines_density)](https://sonarcloud.io/dashboard?id=CharLemAznable_violet)

HTTP容错代理服务. 

使用 [resilience4go](https://github.com/CharLemAznable/resilience4go) 包装代理请求.

支持:
- 舱壁隔离
- 限时
- 限速
- 熔断
- 重试
- 缓存
- 故障恢复

#### 配置样例

请参考: [config_help.toml](https://github.com/CharLemAznable/violet/blob/main/config_help.toml)

#### 数据平面

```violet.NewDataPlane(*violet.Config)``` 

- 实现```http.Handler```接口, 按配置代理http请求
- 支持配置热更新
- 支持Prometheus监控指标收集

#### 控制平面

```violet.NewCtrlPlane(violet.DataPlane)```

- 实现```http.Handler```接口
  - ```/config```: 响应返回当前配置
  - ```/metrics```: 响应返回Prometheus指标数据
  - ```/circuitbreaker/disable```: 停用指定Endpoint的熔断器, 使用url-query或post-form参数```endpoint=xxx```指定Endpoint
  - ```/circuitbreaker/force-open```: 强制开启指定Endpoint的熔断器, 使用url-query或post-form参数```endpoint=xxx```指定Endpoint
  - ```/circuitbreaker/close```: 关闭指定Endpoint的熔断器, 使用url-query或post-form参数```endpoint=xxx```指定Endpoint
  - ```/circuitbreaker/state```: 查询指定Endpoint的熔断器的当前状态, 使用url-query或post-form参数```endpoint=xxx```指定Endpoint
