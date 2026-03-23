---
title: 'Markdown 样式指南'
description: '这里是在 Astro 中编写 Markdown 内容时可以使用的一些基本 Markdown 语法的示例。'
pubDate: 'Jun 19 2024'
---

这里是在 Astro 中编写 Markdown 内容时可以使用的一些基本 Markdown 语法的示例。

## 标题

以下 HTML `<h1>`—`<h6>` 元素代表六个级别的章节标题。`<h1>` 是最高级别，而 `<h6>` 是最低级别。

# 一级标题 (H1)

## 二级标题 (H2)

### 三级标题 (H3)

#### 四级标题 (H4)

##### 五级标题 (H5)

###### 六级标题 (H6)

## 段落

这是一个段落。段落由一个或多个连续的文本行组成，由一个或多个空行隔开。在 Markdown 中，普通的段落不应该用空格或制表符缩进。

你可以通过简单的文本输入来创建段落，Markdown 会自动将其转换为 HTML 中的 `<p>` 标签。

## 图片

### 语法

```markdown
![替代文本](./图片/的/绝对/或/相对/路径)
```

### 输出

## 引用 (Blockquotes)

<blockquote> 元素代表从另一个来源引用的内容。

### 不带出处的引用

#### 语法

```markdown
> 这是引用内容的一个例子。
> **注意** 你可以在引用中使用 _Markdown 语法_。
```

#### 输出

> 这是引用内容的一个例子。  
> **Note** 你可以在引用中使用 _Markdown 语法_。

### 带出处的引用

#### 语法

```markdown
> 不要通过共享内存来通信，而要通过通信来共享内存。<br>
> — <cite>Rob Pike[^1]</cite>
```

#### 输出

> 不要通过共享内存来通信，而要通过通信来共享内存。<br>
> — <cite>Rob Pike[^1]</cite>

[^1]: 以上引用摘自 Rob Pike 在 2015 年 11 月 18 日 Gopherfest 期间的[演讲](https://www.youtube.com/watch?v=PAAkCSZUG1c)。

## 表格

### 语法

```markdown
| 斜体 | 加粗 | 代码 |
| --------- | -------- | ------ |
| _斜体_ | **加粗** | `代码` |
```

### 输出

| 斜体 | 加粗 | 代码 |
| --------- | -------- | ------ |
| _斜体_ | **加粗** | `代码` |

## 代码块

### 语法

我们可以使用 3 个反引号 ``` 在新行中编写代码片段，并在新行中以 3 个反引号结束。要突出显示特定语言的语法，请在第一个 3 个反引号后编写语言名称，例如 html、javascript、css、markdown、typescript、bash。

````markdown
```html
<!doctype html>
<html lang="zh-cn">
  <head>
    <meta charset="utf-8" />
    <title>HTML5 文档示例</title>
  </head>
  <body>
    <p>测试</p>
  </body>
</html>
```
````

### 输出

```html
<!doctype html>
<html lang="zh-cn">
  <head>
    <meta charset="utf-8" />
    <title>HTML5 文档示例</title>
  </head>
  <body>
    <p>测试</p>
  </body>
</html>
```

## 列表类型

### 有序列表

#### 语法

```markdown
1. 第一项
2. 第二项
3. 第三项
```

#### 输出

1. 第一项
2. 第二项
3. 第三项

### 无序列表

#### 语法

```markdown
- 列表项
- 另一项
- 还有一项
```

#### 输出

- 列表项
- 另一项
- 还有一项

### 嵌套列表

#### 语法

```markdown
- 水果
  - 苹果
  - 橙子
  - 香蕉
- 乳制品
  - 牛奶
  - 奶酪
```

#### 输出

- 水果
  - 苹果
  - 橙子
  - 香蕉
- 乳制品
  - 牛奶
  - 奶酪

## 其他元素 — abbr, sub, sup, kbd, mark

### 语法

```markdown
<abbr title="Graphics Interchange Format">GIF</abbr> 是一种位图图像格式。

H<sub>2</sub>O

X<sup>n</sup> + Y<sup>n</sup> = Z<sup>n</sup>

按下 <kbd>CTRL</kbd> + <kbd>ALT</kbd> + <kbd>Delete</kbd> 结束会话。

大多数 <mark>蝾螈</mark> 是夜行动物，捕食昆虫、蠕虫和其他小生物。
```

### 输出

<abbr title="Graphics Interchange Format">GIF</abbr> 是一种位图图像格式。

H<sub>2</sub>O

X<sup>n</sup> + Y<sup>n</sup> = Z<sup>n</sup>

按下 <kbd>CTRL</kbd> + <kbd>ALT</kbd> + <kbd>Delete</kbd> 结束会话。

大多数 <mark>蝾螈</mark> 是夜行动物，捕食昆虫、蠕虫和其他小生物。
