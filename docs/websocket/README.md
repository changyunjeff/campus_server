# WebSocket通信文档

## 1. 连接建立

### 连接地址
ws://{host}/ws?user_id={user_id}

### 参数说明
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 是 | 用户ID |

### 连接示例

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?user_id=12345');
ws.onopen = () => {
    console.log('连接已建立');
};
ws.onclose = () => {
    console.log('连接已关闭');
};
ws.onerror = (error) => {
    console.error('WebSocket错误:', error);
};
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('收到消息:', message);
};
```

## 2. 消息格式

### 2.1 基础消息结构

```javascript
interface Message {
    type: string; // 消息类型
    content: any; // 消息内容
    from?: string; // 发送者ID (发送时可选)
    to: string; // 接收者ID
    createdAt?: Date; // 创建时间 (发送时可选)
    extra?: { // 额外信息
    postId?: string; // 动态ID
    commentId?: string; // 评论ID
    actionType?: string; // 动作类型
    url?: string; // 相关链接
}
```

### 2.2 消息类型
| 类型 | 说明 | 使用场景 |
|------|------|----------|
| chat | 聊天消息 | 用户之间的私聊 |
| like | 点赞通知 | 动态被点赞时 |
| collect | 收藏通知 | 动态被收藏时 |
| comment | 评论通知 | 动态收到新评论时 |
| mention | @通知 | 用户在动态或评论中被@时 |
| system | 系统消息 | 系统通知 |

## 3. 消息发送示例

### 3.1 私聊消息

```javascript
ws.send(JSON.stringify({
    type: 'chat',
    content: '你好!',
    to: 'user_123'
}));
```

### 3.2 点赞通知

```javascript
ws.send(JSON.stringify({
    type: 'like',
    to: 'post_author_id',
    content: '有新的点赞',
    extra: {
        postId: 'post_123',
        actionType: 'like' // 或 'unlike'
    }
}));
```

### 3.3 收藏通知

```javascript
ws.send(JSON.stringify({
    type: 'collect',
    to: 'post_author_id',
    content: '有新的收藏',
    extra: {
        postId: 'post_123',
        actionType: 'collect' // 或 'uncollect'
    }
}));
```

### 3.4 评论通知

```javascript
ws.send(JSON.stringify({
    type: 'comment',
    to: 'post_author_id',
    content: '有新的评论',
    extra: {
        postId: 'post_123',
        commentId: 'comment_456',
        url: '/post/123#comment456'
    }
}));
```

### 3.5 @通知

```javascript
ws.send(JSON.stringify({
    type: 'mention',
    to: 'mentioned_user_id',
    content: '有人@了你',
    extra: {
        postId: 'post_123',
        commentId: 'comment_456',
        url: '/post/123#comment456'
    }
}));
```

## 4. 心跳机制

为保持连接活跃，客户端需要定期发送心跳包：

```javascript
// 每30秒发送一次心跳
setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: 'ping'
        }));
    }
}, 30000);
```

## 5. 工具类实现

推荐使用以下工具类来管理WebSocket连接：

```javascript
// 工具类实现
class WebSocketClient {
    constructor(userId) {
        this.userId = userId;
        this.connect();
    }
    connect() {
        this.ws = new WebSocket(ws://localhost:8080/ws?user_id=${this.userId});
        this.initEventHandlers();
        this.startHeartbeat();
    }
    initEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket连接已建立');
        };
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };
        this.ws.onclose = () => {
            console.log('WebSocket连接已关闭');
            this.stopHeartbeat();
            setTimeout(() => this.connect(), 5000);
        };
        this.ws.onerror = (error) => {
            console.error('WebSocket错误:', error);
        };
    }
    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            this.sendMessage({ type: 'ping' });
        }, 30000);
    }
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
        }
    }
    async sendMessage(message) {
        if (this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket未连接');
        }
    this.ws.send(JSON.stringify(message));
    }
    handleMessage(message) {
        switch (message.type) {
        case 'chat':
            console.log('收到聊天消息:', message);
            break;
        case 'like':
            console.log('收到点赞通知:', message);
            break;
        case 'collect':
            console.log('收到收藏通知:', message);
            break;
        case 'comment':
            console.log('收到评论通知:', message);
            break;
        case 'mention':
            console.log('收到@通知:', message);
            break;
        default:
            console.log('收到其他类型消息:', message);
        }
    }
}
```

## 6. 注意事项

1. 建立连接时必须提供user_id参数
2. 发送消息前确保WebSocket连接状态为OPEN
3. 所有消息必须包含type字段
4. 发送私聊消息时必须指定to字段
5. 建议实现自动重连机制
6. 注意处理网络异常情况
7. 消息内容建议做长度限制

## 7. 错误码说明

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 400 | 缺少user_id参数 | 检查连接URL是否包含user_id |
| 401 | 未授权 | 检查用户登录状态 |
| 1000 | 正常关闭 | 可以重新连接 |
| 1006 | 异常关闭 | 稍后重试 |

如有任何问题，请联系后端开发人员。
