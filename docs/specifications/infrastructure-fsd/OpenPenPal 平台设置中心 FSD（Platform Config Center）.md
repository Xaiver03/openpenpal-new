## **一、模块定位**

  

平台设置中心是 OpenPenPal 后台管理的核心模块，支持跨城市/学校/模块的全局配置项管理与资源控制，包括：

- 城市/学校活动节奏管理
    
- 信封设计模板下发
    
- 条码策略与定价逻辑调整
    
- 开启/关闭功能模块（AI、漂流信、未来信等）
    
- 配置公告与默认推送文案
    

  

该模块仅面向「管理员（admin）」和部分「四级信使（messenger4）」开放。

---

## **二、配置类型一览**

|**配置类型**|**示例内容**|**作用范围**|
|---|---|---|
|📍 城市活动设置|城市征稿开启时间、默认信封编号、奖励金额等|某城市或全局|
|🧾 信封系统设置|信封价格（统一或浮动）、条码自动生成规则、是否允许自打印|全局|
|🧠 AI 功能开关|是否开启 AI 匹配、AI回信节奏（最长间隔、每日推送上限）|某学校或全局|
|📨 功能模块开关|漂流信是否开启、公开信功能是否开放、未来信是否对本校可用|某学校/全局|
|📢 公告与模板文案|公告标题+内容、回信提示模版、邀请信模版（AI笔友初次来信等）|多语种、支持变量|
|🔒 审核与风控设置|敏感词配置、匿名信限制频率、图像上传大小/类型限制|全局|
|💬 默认激励设置|被采纳信封奖励金额、信使成长积分公式、笔友匹配成功后的系统提示语设定|学校/城市级别|

---

## **三、配置结构设计（建议存储结构）**

```
PlatformConfig {
  key: string;                  // 唯一键名，如: ai.enabled
  scope: "global" | "city" | "school";
  scope_id?: string;            // 如 city_id: "HZ", school_code: "PK"
  value: string | number | boolean | object;
  updated_at: datetime;
  updated_by: string;           // 管理员ID
}
```

---

## **四、设置示例（关键字段）**

|**Key 名称**|**默认值**|**类型**|**描述**|
|---|---|---|---|
|ai.enabled|true|boolean|是否开启 AI 笔友模块|
|ai.reply_interval_days|3|number|AI 每封信间隔最短天数|
|envelope.price_cents|300|number|每个信封价格（单位分）|
|barcode.allow_self_print|true|boolean|是否允许用户下载条码自行打印|
|letter.max_anonymous_week|3|number|每周最多匿名信上限|
|school.PK.features.drifting_enabled|false|boolean|是否开启 PK 学校的漂流信功能|
|incentive.envelope_accepted_reward|200|number|被采纳信封奖励金额（单位元）|

---

## **五、配置管理后台建议功能**

- ✅ 搜索与筛选：按关键词、作用域、更新时间筛选
    
- 📝 快速修改：点击键值对可直接编辑并保存
    
- 🧭 作用域切换：全局/城市/学校设置并行管理，层级继承（学校可覆盖城市设置）
    
- 🔒 权限分离：
    
    - 管理员可改所有配置
        
    - 四级信使仅能修改所在城市及下属学校的活动/功能策略
        
    

---

## **六、配置获取接口（用于前端模块判断）**

  

### **6.1 查询全局配置项（前端启动时）**

  

**GET** /api/config/global

  

返回示例：

```
{
  "ai.enabled": true,
  "envelope.price_cents": 300,
  "barcode.allow_self_print": true
}
```

### **6.2 查询当前学校可用功能**

  

**GET** /api/config/school?code=PK

  

返回：

```
{
  "features": {
    "drifting_enabled": false,
    "public_letters": true,
    "future_letters": true
  },
  "ai.reply_interval_days": 2
}
```

---

## **七、安全控制与版本机制**

|**控制点**|**说明**|
|---|---|
|键名前缀命名规范|避免修改错误，配置 key 统一命名如 ai.xxx、school.xx|
|配置操作记录日志|所有变更保留修改记录（修改人、时间、前后对比）|
|临时冻结机制|特定功能可通过 config 实时关闭（如 drifting.enabled = false）|
|快照与回滚功能（未来）|支持配置快照回滚，防误操作|

---

是否继续输出 **数据统计系统（Data Analytics System）** 或者你有更希望优先整理的其他基建模块？✅