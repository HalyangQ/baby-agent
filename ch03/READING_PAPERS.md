# ch03 论文学习清单

本文件整理 `ch03` 相关的论文与技术报告，重点围绕两类模型：专门推理模型，以及通用但推理增强的模型。

## 1) 怎么使用这份清单
一句总结：不要把这些论文当作“背诵材料”，而要把它们当成理解当前推理模型演化路线的地图。

详细内容：
- 这份清单分成两条线来读：一条是专门推理模型，一条是通用但推理增强模型。
- 前一条更偏 reasoning-first 的训练与推理范式，后一条更偏 agent、tool use 和通用模型的推理增强。
- 阅读时不要只看 benchmark 结论，更要看它解决了什么问题、用了什么方法、为什么会成为后续模型路线的基础。

## 2) 专门推理模型这条线
一句总结：如果你想理解今天的 reasoning model 是怎么来的，最值得读的是 `CoT -> Self-Consistency -> STaR -> Let’s Verify -> DeepSeek-R1` 这条链路。

详细内容：
- **Chain-of-Thought Prompting Elicits Reasoning in Large Language Models**  
  这是最基础的一篇 reasoning 入门论文。它回答的是，为什么让模型显式写出中间推理步骤，会显著提升多步任务表现。你可以把它理解成今天大量 reasoning 实践的起点。  
  链接：https://arxiv.org/abs/2201.11903
- **Self-Consistency Improves Chain of Thought Reasoning in Language Models**  
  这篇的重要性在于告诉你，推理能力不只是“有没有 CoT”，还和解码策略强相关。它通过采样多条 reasoning path 再做一致性选择，说明更好的推理结果也可能来自更好的 search / decoding。  
  链接：https://arxiv.org/abs/2203.11171
- **STaR: Bootstrapping Reasoning With Reasoning**  
  这篇很关键，因为它开始把“生成 rationale，再反过来训练自己”的路径讲清楚了。很多后来 reasoning model 的训练思路，都能看到这条路线的影子。  
  链接：https://arxiv.org/abs/2203.14465
- **Let’s Verify Step by Step**  
  这篇非常值得重点读。它代表了“过程监督”路线，也就是不只监督 final answer 对不对，而是去监督中间推理步骤。这对理解 `o1/o3/GPT-5` 这类 reasoning model 的工程思想特别重要。  
  链接：https://arxiv.org/abs/2305.20050
- **DeepSeek-R1: Incentivizing Reasoning Capability in LLMs via Reinforcement Learning**  
  这是当前 reasoning model 学习里非常值得读的一篇技术报告。它更接近今天“推理模型工程化”的现实形态：强化学习、reasoning traces、distillation、open-weight。  
  链接：https://arxiv.org/abs/2501.12948
- **OpenAI o1 System Card**  
  这不是传统学术论文，但非常值得读。它能帮助你理解现代闭源 reasoning model 在产品、评测、安全以及 chain-of-thought 暴露策略上的真实取舍。  
  链接：https://openai.com/index/openai-o1-system-card/

## 3) 通用但推理增强模型这条线
一句总结：这条线更关心的不是“做一个专门推理模型”，而是“如何让通用模型在复杂任务里表现出更强的 reasoning 行为”。

详细内容：
- **ReAct: Synergizing Reasoning and Acting in Language Models**  
  这篇是 agent 学习里的必读论文。它不是纯 reasoning model 训练论文，但它解释了为什么通用模型通过 `Reason -> Act -> Observe` 也能表现出很强的推理行为。  
  链接：https://arxiv.org/abs/2210.03629
- **Toolformer: Language Models Can Teach Themselves to Use Tools**  
  这篇代表另一条非常重要的路线：推理增强不一定来自“想得更深”，也可能来自“更会调用外部工具”。它很适合和 `ch02/ch03` 连起来理解。  
  链接：https://arxiv.org/abs/2302.04761
- **Quiet-STaR: Language Models Can Teach Themselves to Think Before Speaking**  
  这篇很值得看，因为它更接近“通用模型内部如何长出 think-before-speak 行为”，而不是做成一个单独 reasoning-only 产品线。  
  链接：https://arxiv.org/abs/2403.09629

## 4) 建立全景地图的综述
一句总结：如果你想先看全景图，再进具体论文，先读综述会更省力。

详细内容：
- **Reasoning with Language Model Prompting: A Survey**  
  这篇综述很适合当地图。它不是最新，但足够帮你建立 CoT、self-consistency、verification、tool use、planning 这些分支之间的关系。  
  链接：https://arxiv.org/abs/2212.09597

## 5) 建议阅读顺序
一句总结：先理解“推理怎么被诱发出来”，再理解“推理怎么和 acting/tool use 结合”，最后再看“推理怎么被训练、验证和工程化”。

详细内容：
1. `Chain-of-Thought Prompting Elicits Reasoning in Large Language Models`
2. `Self-Consistency Improves Chain of Thought Reasoning in Language Models`
3. `ReAct: Synergizing Reasoning and Acting in Language Models`
4. `Toolformer: Language Models Can Teach Themselves to Use Tools`
5. `STaR: Bootstrapping Reasoning With Reasoning`
6. `Let’s Verify Step by Step`
7. `DeepSeek-R1: Incentivizing Reasoning Capability in LLMs via Reinforcement Learning`
8. `OpenAI o1 System Card`
9. `Quiet-STaR: Language Models Can Teach Themselves to Think Before Speaking`
10. `Reasoning with Language Model Prompting: A Survey`

## 6) 读这些论文时要重点看什么
一句总结：不要只看“模型更强了”，而要看“它到底改变了推理链路里的哪一层”。

详细内容：
- 看它解决的问题是什么：是 prompting、search、verification、tool use，还是训练范式。
- 看它的改进发生在哪一层：输入提示、解码策略、后训练、强化学习，还是运行时控制。
- 看它为什么重要：它是开创一条新路线，还是把已有路线真正做到了工程可用。
- 看它和你当前项目的关系：`ch03` 这章最相关的是 reasoning 暴露、可观察性、以及 reasoning 与 acting 的分层理解。

## 7) 一个现实提醒
一句总结：闭源的“通用但推理增强模型”通常没有完整公开训练论文，所以学习时要接受材料类型不对称。

详细内容：
- 像 `Claude Sonnet 4.6 / Opus 4.6` 这类模型，通常没有像 `DeepSeek-R1` 那样完整公开的训练论文。
- 这时你能依赖的更可靠材料往往是 system card、model report、API 文档和安全报告。
- 所以不要期待所有前沿模型都能用“读一篇技术论文”来完全理解；很多时候你学到的是产品层和系统层的实现线索，而不是完整训练细节。

## 8) 闭源模型没有完整论文时，应该读什么
一句总结：像 `Claude Sonnet 4.6 / Opus 4.6` 这类模型，通常没有像 `DeepSeek-R1` 那样完整公开的训练论文，这时更应该读发布博客、system card、model report 和帮助文档。

详细内容：
- **Anthropic 官方最值得看的材料：**  
  对 `Claude Sonnet 4.6 / Opus 4.6` 来说，最值得看的官方材料包括 Sonnet 4.6 的发布博客、Transparency Hub / Model Report、System Cards，以及 `extended thinking` 的帮助文档。这些材料虽然不会公开到底层训练架构细节，但足够让你理解产品定位、thinking 能力暴露方式、agentic 能力和长上下文使用场景。偏 system card / model report，不是完整技术论文
  - Sonnet 4.6 发布博客
    重点讲能力、长上下文、agentic coding、adaptive thinking / extended thinking。
    来源：https://www.anthropic.com/news/claude-sonnet-4-6
  - Transparency Hub / Model Report
    这里有最官方的模型描述。Anthropic明确把 Claude Opus 4.6 描述为 hybrid reasoning large language model。
    来源：https://www.anthropic.com/transparency/model-report
  - System Cards
    这里能看到 Sonnet 4.6、Opus 4.6 的系统卡入口。
    来源：https://www.anthropic.com/system-cards
  - Extended Thinking 帮助文档
    这个不是架构文档，但能帮助你理解 Anthropic 产品层是怎么暴露 thinking 的。
    来源：https://support.claude.com/en/articles/10574485-using-extended-thinking
- **官方明确公开了什么：**  
  Anthropic 已经明确公开过 `Claude Opus 4.6` 是 `hybrid reasoning large language model`，也明确说明 Claude 4.6 路线支持 `adaptive thinking` 和 `extended thinking`。这能帮助你理解 Claude 4.6 属于通用模型上叠加 reasoning 行为控制的路线，而不是单纯的传统 chat model。
- **官方没有公开什么：**  
  目前 Anthropic 没有公开到底层网络结构、是否是 MoE、参数规模、router 细节，以及 thinking 到底是单模型模式切换还是多模型系统。因此如果你想进一步知道“底层究竟怎么做”，就不能把社区猜测当成事实。
- **社区猜测能不能看：**  
  可以看，但只能当工程直觉材料，不能当结论。社区常见猜法包括：Claude 可能存在 router / effort control、可能是 hybrid reasoning 路线、可能采用 MoE 或部分激活参数架构，也有人用吞吐量和 token 速度反推 active parameters。这类内容的价值在于帮助你形成问题意识，而不是直接得到可靠答案。一个例子是这类基于吞吐量的逆向估计：  
  https://agent-wars.com/news/2026-03-13-estimating-the-size-of-claude-opus-4-5-4-6
- **更稳的学习方式：**  
  对闭源模型，建议优先学“系统和产品层”的可公开信息，而不是过早沉迷于未经证实的底层架构猜测。也就是说，先理解它的 thinking 模式、上下文能力、工具使用方式、延迟成本取向和安全约束，再去看社区的 reverse engineering 分析。
