import type { Topic, Question, Lesson, QuizAttempt } from '@/types';

const lessons_llms: Lesson[] = [
  {
    id: 'les-llm-1',
    topicId: 'top-llms',
    title: 'What Is a Large Language Model?',
    summary: 'The mental model: tokens, next-token prediction, and why scale changes behavior.',
    position: 1,
    durationMin: 8,
    status: 'published',
    body: `A large language model (LLM) is a neural network trained to predict the next token in a sequence. From that single objective — **next-token prediction** — a surprising range of capabilities emerge: translation, summarization, code generation, reasoning, and more.

## Tokens, not words

LLMs process text as **tokens**: chunks of characters that roughly correspond to word pieces. The word "unbelievable" might split into ["un", "believ", "able"]. A token vocabulary typically holds 50,000–200,000 entries.

\`\`\`
text:      "The model learns"
tokens:    [464, 2061, 8372]
\`\`\`

The model outputs a probability distribution over the entire vocabulary for the *next* token, then samples or picks one.

## Why scale matters

As you scale up parameters, data, and compute, models don't just get *better* — they gain **new capabilities**. This is called *emergent capability*. Few-shot in-context learning, chain-of-thought reasoning, and instruction following all appear above certain scale thresholds.

> More parameters + more data + more compute = qualitatively different behavior, not just incremental gains.

## The training pipeline

Most modern LLMs go through three stages:

1. **Pre-training** — predict the next token on a massive text corpus (trillions of tokens).
2. **Supervised fine-tuning (SFT)** — train on high-quality instruction/response pairs.
3. **RLHF / preference tuning** — align the model to human preferences using reinforcement learning or direct preference optimization (DPO).

## Key takeaway

An LLM is a next-token predictor trained at scale. Everything you build on top — chat, agents, RAG, tools — is orchestration *around* this core mechanism. Understanding this helps you reason about why LLMs fail and how to design around their limits.`,
  },
  {
    id: 'les-llm-2',
    topicId: 'top-llms',
    title: 'Context Windows and Token Limits',
    summary: 'How context length affects cost, latency, and what your app can actually do.',
    position: 2,
    durationMin: 7,
    status: 'published',
    body: `The **context window** is the maximum number of tokens a model can attend to in a single request — input plus output. A model with an 8K context can't "see" more than 8,000 tokens at once.

## What fits in a context window

Rough estimates:

- 1 token ≈ 4 characters in English
- 750 tokens ≈ 1,000 words
- A 32K window ≈ a 24-page document
- A 128K window ≈ a 300-page book

## Cost and latency scale with tokens

API pricing is per-token. Attention computation is roughly O(n²) in sequence length for standard transformers. Doubling context often more than doubles latency and cost.

| Context size | Relative latency | Relative cost |
|---|---|---|
| 4K | 1x | 1x |
| 32K | ~3x | ~8x |
| 128K | ~10x | ~32x |

## Practical guidance

- **Stuff only what you need.** Long context doesn't mean you should put everything in it. Irrelevant context degrades accuracy.
- **Lost-in-the-middle** is real: models attend less reliably to information in the middle of long contexts. Put important instructions at the *start* and *end*.
- **Cache prefix** when you can. Anthropic and OpenAI both offer prompt caching that dramatically cuts cost on repeated prefixes.

## Key takeaway

Bigger context windows aren't a free lunch. They cost more, run slower, and can hurt accuracy. Treat context as a scarce resource: put the right things in, leave noise out.`,
  },
  {
    id: 'les-llm-3',
    topicId: 'top-llms',
    title: 'Sampling: Temperature, Top-p, and Why They Matter',
    summary: 'Turning probabilities into text, and how to tune it for your use case.',
    position: 3,
    durationMin: 6,
    status: 'published',
    body: `After the model produces a probability distribution over the vocabulary, you have to **pick** a token. How you pick has a big effect on output quality.

## Temperature

Temperature rescales the logits before softmax:

- **temperature = 0** → always pick the highest-probability token (deterministic, "greedy")
- **temperature = 1** → sample proportional to the model's raw distribution
- **temperature > 1** → flatten the distribution, more random

## Top-p (nucleus) sampling

Instead of sampling from all tokens, sample only from the smallest set of tokens whose cumulative probability exceeds *p* (e.g. 0.9). This cuts off the long tail of absurd low-probability tokens.

## When to use what

- **Code generation / data extraction** → temperature 0. You want determinism and the most likely answer.
- **Creative writing / brainstorming** → temperature 0.7–0.9. Diversity helps.
- **Customer-facing chat** → temperature 0.4–0.6 with top-p 0.9. Some variety, but stay coherent.

> A common mistake: cranking temperature to 1.5 hoping for "smarter" answers. You get *random*, not *smart*.

## Key takeaway

Sampling parameters control the *shape* of outputs, not the model's capability. Pick values that match your task: deterministic for extraction, moderate for chat, high for brainstorming.`,
  },
];

const questions_llms: Question[] = [
  {
    id: 'q-llm-1',
    topicId: 'top-llms',
    text: 'What is the core training objective of a large language model during pre-training?',
    explanation: 'LLMs are trained with next-token prediction: maximize the probability of the correct next token given the preceding context.',
    options: [
      { id: 'o1', text: 'Classify sentences into categories', isCorrect: false },
      { id: 'o2', text: 'Predict the next token in a sequence', isCorrect: true },
      { id: 'o3', text: 'Minimize the length of generated text', isCorrect: false },
      { id: 'o4', text: 'Translate between language pairs', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-2',
    topicId: 'top-llms',
    text: 'Roughly how many characters does one token correspond to in English text?',
    explanation: 'A useful rule of thumb is 1 token ≈ 4 characters, or about 750 tokens per 1,000 words.',
    options: [
      { id: 'o1', text: 'About 1 character', isCorrect: false },
      { id: 'o2', text: 'About 4 characters', isCorrect: true },
      { id: 'o3', text: 'About 10 characters', isCorrect: false },
      { id: 'o4', text: 'About 100 characters', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-3',
    topicId: 'top-llms',
    text: 'Which stage of training teaches the model to follow instructions in a chat format?',
    explanation: 'Supervised fine-tuning (SFT) trains on instruction/response pairs so the model follows instructions; RLHF then aligns it to preferences.',
    options: [
      { id: 'o1', text: 'Pre-training on web text', isCorrect: false },
      { id: 'o2', text: 'Supervised fine-tuning (SFT)', isCorrect: true },
      { id: 'o3', text: 'Tokenization', isCorrect: false },
      { id: 'o4', text: 'Vocabulary expansion', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-4',
    topicId: 'top-llms',
    text: 'What does a temperature of 0 do during sampling?',
    explanation: 'Temperature 0 makes sampling greedy: the highest-probability token is always selected, giving deterministic output.',
    options: [
      { id: 'o1', text: 'Randomizes outputs maximally', isCorrect: false },
      { id: 'o2', text: 'Always picks the highest-probability token', isCorrect: true },
      { id: 'o3', text: 'Disables the model entirely', isCorrect: false },
      { id: 'o4', text: 'Doubles the context window', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-5',
    topicId: 'top-llms',
    text: 'What is the "lost-in-the-middle" phenomenon?',
    explanation: 'Models attend less reliably to information in the middle of long contexts; placing key facts at the start or end improves recall.',
    options: [
      { id: 'o1', text: 'Models forget their system prompt', isCorrect: false },
      { id: 'o2', text: 'Models attend less reliably to information in the middle of long contexts', isCorrect: true },
      { id: 'o3', text: 'Output gets cut off at 50% of max tokens', isCorrect: false },
      { id: 'o4', text: 'The middle layers of the network degrade first', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-6',
    topicId: 'top-llms',
    text: 'Why does doubling the context length often more than double latency?',
    explanation: 'Standard transformer attention is roughly O(n²) in sequence length, so doubling context can quadruple the core attention cost.',
    options: [
      { id: 'o1', text: 'Because API providers charge a minimum fee', isCorrect: false },
      { id: 'o2', text: 'Because attention computation scales roughly quadratically with sequence length', isCorrect: true },
      { id: 'o3', text: 'Because longer contexts always use a slower model variant', isCorrect: false },
      { id: 'o4', text: 'It does not — latency is constant regardless of context', isCorrect: false },
    ],
  },
  {
    id: 'q-llm-7',
    topicId: 'top-llms',
    text: 'What does top-p (nucleus) sampling do?',
    explanation: 'Top-p sampling restricts the candidate set to the smallest group of tokens whose cumulative probability exceeds p, cutting off the long tail.',
    options: [
      { id: 'o1', text: 'Picks the top p tokens regardless of probability', isCorrect: false },
      { id: 'o2', text: 'Samples only from the smallest token set whose cumulative probability exceeds p', isCorrect: true },
      { id: 'o3', text: 'Truncates the context to p tokens', isCorrect: false },
      { id: 'o4', text: 'Sets temperature to p', isCorrect: false },
    ],
  },
];

const lessons_agents: Lesson[] = [
  {
    id: 'les-agt-1',
    topicId: 'top-agents',
    title: 'What Is an AI Agent?',
    summary: 'Beyond single completions: loops, tools, and the model as a decision-maker.',
    position: 1,
    durationMin: 9,
    status: 'published',
    body: `An **AI agent** is a system where a language model decides, in a loop, what to do next — including calling external tools — until it reaches a goal or a stop condition.

## The agent loop

A typical agent runs this cycle:

\`\`\`
1. Read the goal and current state
2. Decide: call a tool, or produce a final answer
3. If a tool was called, append the result to state
4. Repeat until done or max steps reached
\`\`\`

The key difference from a single chat completion is the **loop**. The model isn't just answering once — it's planning, acting, observing, and re-planning.

## Tools ground the model

Without tools, a model can only say what it knows. With tools, it can:

- **Retrieve** real-time or private data (search, database lookups)
- **Compute** exactly (calculator, code interpreter)
- **Act** in the world (send an email, update a ticket, run a query)

Tools turn the LLM from a talker into a doer.

## When *not* to build an agent

Agents add latency, cost, and unpredictability. If a single prompt reliably solves your task, use a single prompt. Reach for an agent when:

- The task needs multiple steps you can't enumerate up front.
- The task needs external data or actions the model can't predict.
- A fixed workflow would break on edge cases the model can handle.

> "Agentic" is not always better. A well-prompted single completion beats a flaky 10-step loop.

## Key takeaway

An agent is an LLM in a loop with tools. The model decides the next step; tools turn decisions into actions. Use agents when the path to the answer isn't known in advance.`,
  },
  {
    id: 'les-agt-2',
    topicId: 'top-agents',
    title: 'Tool Calling and Function Definitions',
    summary: 'How to expose tools to a model and parse its calls safely.',
    position: 2,
    durationMin: 8,
    status: 'published',
    body: `Modern LLM APIs let you declare **tools** (functions) with a JSON schema. The model can then choose to emit a structured call to one of them instead of plain text.

## Anatomy of a tool definition

\`\`\`json
{
  "name": "search_docs",
  "description": "Search internal documentation by query.",
  "parameters": {
    "type": "object",
    "properties": {
      "query": { "type": "string", "description": "The search query" },
      "limit": { "type": "integer", "minimum": 1, "maximum": 20 }
    },
    "required": ["query"]
  }
}
\`\`\`

The model sees this schema and decides whether to call the tool, with what arguments.

## The round trip

1. You send the prompt + the tool definitions.
2. The model responds with a tool call (name + arguments) instead of text.
3. **You execute the tool** in your code — the model cannot run anything itself.
4. You send the tool result back as a new message.
5. The model continues, either calling another tool or producing a final answer.

## Safety and validation

- **Never trust the model's arguments blindly.** Validate types, ranges, and permissions before executing. The model can — and will — emit malformed or out-of-bounds arguments.
- **Limit the blast radius.** A "delete_file" tool in an agent loop is a footgun. Prefer read tools, and gate writes behind explicit human approval.
- **Description is the API.** The model picks tools based on the description string. Spend time writing clear, specific descriptions; vague descriptions cause wrong-tool calls.

## Key takeaway

Tool calling is a structured contract: you declare functions, the model emits calls, you execute and return results. Validate everything the model sends, and write descriptions as if the model's correctness depends on them — because it does.`,
  },
  {
    id: 'les-agt-3',
    topicId: 'top-agents',
    title: 'Planning, ReAct, and Loop Control',
    summary: 'Patterns for keeping agents productive instead of runaway.',
    position: 3,
    durationMin: 7,
    status: 'published',
    body: `Giving a model tools isn't enough — you need a **control loop** that turns tool calls into progress. Several patterns have emerged.

## ReAct (Reason + Act)

The model alternates between a **Thought** (reasoning about the next step) and an **Action** (a tool call), then observes the result and thinks again.

\`\`\`
Thought: I need the user's recent orders.
Action: get_orders(user_id="u_123")
Observation: [3 orders, last on 2024-11-02]
Thought: The most recent order is from Nov 2; I can answer now.
Final Answer: ...
\`\`\`

## Plan-then-execute

First ask the model to produce a **plan** (a list of steps), then execute each step. This separates reasoning about *what* to do from *doing* it, and lets you check the plan before running expensive or irreversible steps.

## Loop control essentials

- **Max steps.** Always cap iterations (e.g. 10). Without a cap, a confused agent loops forever.
- **Timeouts.** Each tool call needs a timeout. A hung tool hangs the whole agent.
- **Error feedback.** When a tool fails, return the error to the model as the observation. It can often recover.
- **Stop condition.** The model emits a "final answer" or you detect a terminal tool.

## When agents go wrong

- **Tool spiral** — calling the same tool repeatedly with slightly different arguments.
- **Hallucinated tool** — inventing a tool name that wasn't declared.
- **Lost goal** — drifting away from the original task after several steps.

Mitigate with max steps, periodic re-injection of the original goal, and logging every step.

## Key takeaway

A good agent loop has guardrails: a step cap, timeouts, error feedback, and a clear stop condition. The pattern (ReAct, plan-then-execute) matters less than the discipline of the loop.`,
  },
];

const questions_agents: Question[] = [
  {
    id: 'q-agt-1',
    topicId: 'top-agents',
    text: 'What distinguishes an AI agent from a single chat completion?',
    explanation: 'An agent runs the model in a loop, deciding next steps and calling tools, rather than producing one answer.',
    options: [
      { id: 'o1', text: 'It uses a larger model', isCorrect: false },
      { id: 'o2', text: 'It runs the model in a loop with tools to decide next steps', isCorrect: true },
      { id: 'o3', text: 'It always streams the response', isCorrect: false },
      { id: 'o4', text: 'It does not use any prompts', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-2',
    topicId: 'top-agents',
    text: 'Who actually executes a tool call emitted by the model?',
    explanation: 'The model only emits a structured call; your code is responsible for executing the tool and returning the result.',
    options: [
      { id: 'o1', text: 'The model itself, inside the API', isCorrect: false },
      { id: 'o2', text: 'Your application code', isCorrect: true },
      { id: 'o3', text: 'The API provider servers', isCorrect: false },
      { id: 'o4', text: 'No one — tool calls are simulated', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-3',
    topicId: 'top-agents',
    text: 'What is the ReAct pattern?',
    explanation: 'ReAct alternates reasoning (Thought) and action (tool call), observing results before reasoning again.',
    options: [
      { id: 'o1', text: 'Reasoning + acting in alternating steps, observing between them', isCorrect: true },
      { id: 'o2', text: 'A single forward pass with no tools', isCorrect: false },
      { id: 'o3', text: 'Reinforcement learning from tool feedback', isCorrect: false },
      { id: 'o4', text: 'A reduced-attention transformer variant', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-4',
    topicId: 'top-agents',
    text: 'Why must you validate the arguments the model sends to a tool?',
    explanation: 'The model can emit malformed or out-of-bounds arguments; validating types, ranges, and permissions prevents unsafe execution.',
    options: [
      { id: 'o1', text: 'To reduce API latency', isCorrect: false },
      { id: 'o2', text: 'Because the model can emit malformed or unsafe arguments', isCorrect: true },
      { id: 'o3', text: 'Because JSON is illegal in some regions', isCorrect: false },
      { id: 'o4', text: 'Validation is optional and rarely needed', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-5',
    topicId: 'top-agents',
    text: 'What is a tool spiral failure mode?',
    explanation: 'A tool spiral is when an agent repeatedly calls the same tool with slightly different arguments without making progress.',
    options: [
      { id: 'o1', text: 'Calling the same tool repeatedly with slightly different arguments', isCorrect: true },
      { id: 'o2', text: 'Running out of API quota', isCorrect: false },
      { id: 'o3', text: 'The model refusing to use tools', isCorrect: false },
      { id: 'o4', text: 'A tool that returns circular JSON', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-6',
    topicId: 'top-agents',
    text: 'Which is an essential safeguard for any agent loop?',
    explanation: 'A maximum number of steps prevents runaway loops; every agent should have a cap, timeouts, and a stop condition.',
    options: [
      { id: 'o1', text: 'A maximum number of steps', isCorrect: true },
      { id: 'o2', text: 'A temperature above 1.0', isCorrect: false },
      { id: 'o3', text: 'Disabling all error messages', isCorrect: false },
      { id: 'o4', text: 'Using only one tool at a time', isCorrect: false },
    ],
  },
  {
    id: 'q-agt-7',
    topicId: 'top-agents',
    text: 'In plan-then-execute, why separate planning from execution?',
    explanation: 'Separating planning lets you inspect or validate the plan before running expensive or irreversible steps.',
    options: [
      { id: 'o1', text: 'It makes the model faster', isCorrect: false },
      { id: 'o2', text: 'It lets you check the plan before running costly or irreversible steps', isCorrect: true },
      { id: 'o3', text: 'It avoids paying API costs entirely', isCorrect: false },
      { id: 'o4', text: 'It is required by the OpenAI API', isCorrect: false },
    ],
  },
];

const lessons_mcp: Lesson[] = [
  {
    id: 'les-mcp-1',
    topicId: 'top-mcp',
    title: 'Introducing the Model Context Protocol',
    summary: 'A standard way for models to reach tools, resources, and prompts.',
    position: 1,
    durationMin: 8,
    status: 'published',
    body: `The **Model Context Protocol (MCP)** is an open standard for connecting language models to external data sources and tools. Think of it as USB-C for AI integrations: one protocol instead of a thousand bespoke connectors.

## The problem MCP solves

Before MCP, every app that wanted to give a model access to (say) a GitHub repo, a Slack workspace, or a database wrote its own integration. Each had its own auth, schema, and invocation conventions. MCP standardizes this exchange.

## Three primitives

MCP servers expose three kinds of things:

1. **Tools** — functions the model can call (same idea as function calling, but standardized).
2. **Resources** — typed data the model can read (files, database rows, API responses).
3. **Prompts** — reusable prompt templates the client can fill in.

## Host, client, server

- The **host** is the user-facing app (an IDE, a chat client).
- The **client** lives inside the host and speaks the MCP protocol.
- The **server** is the integration (a GitHub MCP server, a Postgres MCP server).

A host can connect to many servers at once, giving the model a unified view of many integrations.

## Key takeaway

MCP standardizes the connection between models and the outside world. Tools, resources, and prompts are the primitives; host/client/server is the topology. The value is the standard — write a server once, any MCP-aware client can use it.`,
  },
  {
    id: 'les-mcp-2',
    topicId: 'top-mcp',
    title: 'MCP Transports: stdio and HTTP',
    summary: 'How MCP clients and servers actually talk to each other.',
    position: 2,
    durationMin: 6,
    status: 'published',
    body: `MCP defines the *messages*; a **transport** defines how those messages move between client and server. Two are standard.

## stdio transport

The client launches the server as a **child process** and communicates over standard input/output. JSON-RPC messages are framed line by line.

- **Use it when** the server runs locally on the user's machine, alongside the host app.
- **Pros:** no network, no auth, no port conflicts. Simple to ship.
- **Cons:** server must run on the same host as the client; not suitable for remote integrations.

## HTTP + SSE transport

The server runs as a remote HTTP service. The client sends requests over HTTP and receives server-initiated messages via Server-Sent Events.

- **Use it when** the server is shared, remote, or hosted.
- **Pros:** many clients can connect; survives across sessions.
- **Cons:** you need auth, TLS, and to handle reconnection.

## Choosing

- Local developer tools (a CLI that exposes your repo) → **stdio**.
- A shared company-wide integration (a hosted CRM MCP server) → **HTTP**.

## Key takeaway

Pick the transport by where the server lives: stdio for local per-user servers, HTTP for shared remote servers. The protocol messages stay the same either way.`,
  },
];

const questions_mcp: Question[] = [
  {
    id: 'q-mcp-1',
    topicId: 'top-mcp',
    text: 'What is the Model Context Protocol (MCP)?',
    explanation: 'MCP is an open standard for connecting language models to external data sources and tools, reducing bespoke integrations.',
    options: [
      { id: 'o1', text: 'A proprietary OpenAI API for embeddings', isCorrect: false },
      { id: 'o2', text: 'An open standard for connecting models to tools, resources, and prompts', isCorrect: true },
      { id: 'o3', text: 'A Go framework for building APIs', isCorrect: false },
      { id: 'o4', text: 'A model-quantization algorithm', isCorrect: false },
    ],
  },
  {
    id: 'q-mcp-2',
    topicId: 'top-mcp',
    text: 'Which is NOT one of the three MCP primitives?',
    explanation: 'MCP servers expose tools, resources, and prompts. "Datasets" is not a primitive.',
    options: [
      { id: 'o1', text: 'Tools', isCorrect: false },
      { id: 'o2', text: 'Resources', isCorrect: false },
      { id: 'o3', text: 'Prompts', isCorrect: false },
      { id: 'o4', text: 'Datasets', isCorrect: true },
    ],
  },
  {
    id: 'q-mcp-3',
    topicId: 'top-mcp',
    text: 'In MCP, what is the "host"?',
    explanation: 'The host is the user-facing app (e.g. an IDE or chat client) that contains one or more MCP clients.',
    options: [
      { id: 'o1', text: 'The user-facing application that contains MCP clients', isCorrect: true },
      { id: 'o2', text: 'The remote integration server', isCorrect: false },
      { id: 'o3', text: 'The model provider', isCorrect: false },
      { id: 'o4', text: 'The physical server hardware', isCorrect: false },
    ],
  },
  {
    id: 'q-mcp-4',
    topicId: 'top-mcp',
    text: 'When should you use the stdio transport?',
    explanation: 'stdio is ideal when the server runs locally on the user machine alongside the host, with no network needed.',
    options: [
      { id: 'o1', text: 'When the server runs locally on the user machine alongside the host', isCorrect: true },
      { id: 'o2', text: 'When the server is shared by many users over the network', isCorrect: false },
      { id: 'o3', text: 'When you need TLS and OAuth', isCorrect: false },
      { id: 'o4', text: 'When the host is a web browser only', isCorrect: false },
    ],
  },
  {
    id: 'q-mcp-5',
    topicId: 'top-mcp',
    text: 'Which transport is better for a shared, remote MCP server?',
    explanation: 'HTTP + SSE supports remote, multi-client servers and handles server-initiated messages via Server-Sent Events.',
    options: [
      { id: 'o1', text: 'stdio', isCorrect: false },
      { id: 'o2', text: 'HTTP + SSE', isCorrect: true },
      { id: 'o3', text: 'Raw TCP without framing', isCorrect: false },
      { id: 'o4', text: 'WebSockets over FTP', isCorrect: false },
    ],
  },
  {
    id: 'q-mcp-6',
    topicId: 'top-mcp',
    text: 'What is a key benefit of MCP standardizing integrations?',
    explanation: 'A single MCP server can be used by any MCP-aware client, so you write the integration once.',
    options: [
      { id: 'o1', text: 'It removes the need for any prompts', isCorrect: false },
      { id: 'o2', text: 'Write a server once and any MCP-aware client can use it', isCorrect: true },
      { id: 'o3', text: 'It makes models train themselves', isCorrect: false },
      { id: 'o4', text: 'It eliminates API authentication', isCorrect: false },
    ],
  },
];

const lessons_rag: Lesson[] = [
  {
    id: 'les-rag-1',
    topicId: 'top-rag',
    title: 'Retrieval-Augmented Generation: The Idea',
    summary: 'Give the model the right context at query time instead of relying on memorized facts.',
    position: 1,
    durationMin: 9,
    status: 'published',
    body: `**Retrieval-Augmented Generation (RAG)** combines a search step with a generation step. Instead of asking the model to answer from what it memorized during training, you first *retrieve* relevant documents and then ask the model to answer using those.

## Why RAG?

- **Freshness** — knowledge can be updated without retraining.
- **Private data** — your company's docs aren't in the training set.
- **Citations** — you can show the user where the answer came from.
- **Cheaper than fine-tuning** — updating an index is far cheaper than retraining.

## The basic pipeline

\`\`\`
query → embed → search index → top-k chunks → prompt with chunks → model → answer
\`\`\`

1. Turn the user's query into a vector embedding.
2. Search a vector index for the k most similar chunks.
3. Insert those chunks into the prompt as context.
4. Ask the model to answer using only that context.

## When RAG fails

- **Bad chunking** — chunks that are too big dilute relevance; too small lose context.
- **Bad retrieval** — if the right document isn't in the top-k, the model can't use it.
- **Stuffed context** — dumping 20 mediocre chunks hurts more than 3 great ones.
- **No answer grounding** — if you don't instruct the model to cite, it may hallucinate beyond the context.

## Key takeaway

RAG = retrieve relevant context, then generate. It's the standard pattern for giving models access to private or fresh knowledge. Its quality is dominated by the *retrieval* step, not the model.`,
  },
  {
    id: 'les-rag-2',
    topicId: 'top-rag',
    title: 'Chunking Strategies',
    summary: 'How you split documents determines what retrieval can find.',
    position: 2,
    durationMin: 7,
    status: 'published',
    body: `**Chunking** is how you split documents into pieces that get embedded and indexed. It's the single most under-appreciated lever in RAG quality.

## Size matters

- **Too small** (e.g. single sentences) — each chunk loses context; a chunk may match a query semantically but lack the surrounding facts needed to answer.
- **Too large** (e.g. whole pages) — the embedding averages over too much text; relevance signals get diluted, and you burn context budget.

A common starting point: **400–800 tokens per chunk**.

## Overlap

Add **overlap** (e.g. 50–100 tokens) between adjacent chunks so a fact that straddles a boundary isn't lost. Without overlap, a sentence cut in half becomes two weaker chunks.

## Structure-aware chunking

Naive chunking splits on character count. Better approaches respect structure:

- **Markdown** — split on headings, keep heading hierarchy in each chunk.
- **Code** — split on function/class boundaries.
- **PDFs** — respect columns and page breaks.

## Metadata

Attach metadata to each chunk (source, section, page, date). You can then **filter** before retrieval: "only search chunks from the 2024 handbook" or "only the API reference section."

## Key takeaway

Chunking decides what the retriever can find. Pick a sensible size (400–800 tokens), add overlap, respect document structure, and attach metadata so you can filter as well as search.`,
  },
  {
    id: 'les-rag-3',
    topicId: 'top-rag',
    title: 'Hybrid Search and Reranking',
    summary: "Vector search alone isn't always enough; combining signals beats any single method.",
    position: 3,
    durationMin: 7,
    status: 'published',
    body: `Pure vector similarity is good at *semantic* matches but bad at *exact* matches — names, IDs, error codes, acronyms. **Hybrid search** combines it with keyword (BM25) search to get the best of both.

## The hybrid recipe

1. Run **BM25** (keyword) search → top N candidates.
2. Run **vector** search → top N candidates.
3. Merge the two lists with a weighted score (e.g. 0.3 BM25 + 0.7 vector).
4. **Rerank** the merged top candidates with a cross-encoder.

## Why reranking matters

Vector retrieval is fast but approximate — it's a *bi-encoder* that scores query and document separately. A **cross-encoder** reranker scores query and document *together*, catching fine-grained relevance the bi-encoder misses.

\`\`\`
retrieve 50 candidates (fast, approximate)
  → rerank to top 5 (slower, accurate)
\`\`\`

This retrieve-then-rerank pattern beats either step alone.

## Tuning

- **Top-k for retrieval** — larger pools give the reranker more to work with but cost more.
- **Weight blend** — tune the BM25/vector weight for your corpus. Code and logs benefit from more BM25; prose benefits from more vector.
- **Reranker model** — small cross-encoders (e.g. ms-marco-MiniLM) are cheap and effective.

## Key takeaway

Don't stop at vector search. Combine BM25 and vector retrieval, then rerank the top candidates with a cross-encoder. The rerank step often improves answer quality more than swapping the LLM.`,
  },
];

const questions_rag: Question[] = [
  {
    id: 'q-rag-1',
    topicId: 'top-rag',
    text: 'What is the core idea of retrieval-augmented generation?',
    explanation: 'RAG retrieves relevant documents at query time and conditions the model on them, rather than relying on memorized training knowledge.',
    options: [
      { id: 'o1', text: 'Fine-tune the model on every query', isCorrect: false },
      { id: 'o2', text: 'Retrieve relevant context at query time, then generate using it', isCorrect: true },
      { id: 'o3', text: 'Cache model outputs forever', isCorrect: false },
      { id: 'o4', text: 'Train without any text data', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-2',
    topicId: 'top-rag',
    text: 'Which is a common failure mode caused by bad chunking?',
    explanation: 'Poorly sized chunks either dilute relevance (too big) or lose surrounding context (too small), hurting retrieval quality.',
    options: [
      { id: 'o1', text: 'Chunks too small lose surrounding context needed to answer', isCorrect: true },
      { id: 'o2', text: 'The model refuses to answer any question', isCorrect: false },
      { id: 'o3', text: 'Embeddings become integers', isCorrect: false },
      { id: 'o4', text: 'The vector index deletes itself', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-3',
    topicId: 'top-rag',
    text: 'Why add overlap between chunks?',
    explanation: 'Overlap ensures facts that straddle a chunk boundary are not lost, preserving context across the split.',
    options: [
      { id: 'o1', text: 'To reduce storage cost', isCorrect: false },
      { id: 'o2', text: 'So a fact that straddles a boundary is not lost', isCorrect: true },
      { id: 'o3', text: 'To increase the embedding dimension', isCorrect: false },
      { id: 'o4', text: 'Overlap is required by all vector databases', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-4',
    topicId: 'top-rag',
    text: 'What does hybrid search combine?',
    explanation: 'Hybrid search blends BM25 keyword search with vector similarity search to cover both exact and semantic matches.',
    options: [
      { id: 'o1', text: 'BM25 keyword search and vector similarity search', isCorrect: true },
      { id: 'o2', text: 'Two different LLMs voting', isCorrect: false },
      { id: 'o3', text: 'SQL and GraphQL', isCorrect: false },
      { id: 'o4', text: 'Pre-training and fine-tuning', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-5',
    topicId: 'top-rag',
    text: 'What does a cross-encoder reranker do that a bi-encoder retriever does not?',
    explanation: 'A cross-encoder scores the query and document together, capturing fine-grained relevance the bi-encoder scores separately and misses.',
    options: [
      { id: 'o1', text: 'Scores query and document together for finer relevance', isCorrect: true },
      { id: 'o2', text: 'Generates the final answer', isCorrect: false },
      { id: 'o3', text: 'Compresses the embedding to 1 dimension', isCorrect: false },
      { id: 'o4', text: 'Translates the query to SQL', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-6',
    topicId: 'top-rag',
    text: 'Why attach metadata to chunks?',
    explanation: 'Metadata lets you filter before retrieval (e.g. by source, section, or date), narrowing the search to relevant subsets.',
    options: [
      { id: 'o1', text: 'To increase embedding dimensions', isCorrect: false },
      { id: 'o2', text: 'To allow filtering before retrieval by source, section, or date', isCorrect: true },
      { id: 'o3', text: 'To avoid needing a vector index', isCorrect: false },
      { id: 'o4', text: 'Metadata is purely decorative', isCorrect: false },
    ],
  },
  {
    id: 'q-rag-7',
    topicId: 'top-rag',
    text: 'Which RAG step usually dominates answer quality?',
    explanation: 'Retrieval quality dominates: if the right document is not in the top-k, no model can rescue the answer.',
    options: [
      { id: 'o1', text: 'The retrieval step', isCorrect: true },
      { id: 'o2', text: 'The font used in the UI', isCorrect: false },
      { id: 'o3', text: 'The size of the prompt template', isCorrect: false },
      { id: 'o4', text: 'The order of API headers', isCorrect: false },
    ],
  },
];

const lessons_embeddings: Lesson[] = [
  {
    id: 'les-emb-1',
    topicId: 'top-emb',
    title: 'What Embeddings Actually Are',
    summary: 'Vectors that capture meaning, and how they are produced.',
    position: 1,
    durationMin: 8,
    status: 'published',
    body: `An **embedding** is a vector of numbers (often 384, 768, or 1536 dimensions) that represents a piece of content — a sentence, a chunk of text, an image. **Similar content gets similar vectors.**

## Where they come from

Embedding models are trained so that *semantically similar* inputs end up close in vector space. A common training objective is **contrastive learning**: pull related pairs together, push unrelated pairs apart.

\`\`\`
"how do I reset my password"  →  [0.21, -0.04, 0.88, ...]
"password reset instructions"  →  [0.19, -0.02, 0.85, ...]   (close)
"today's weather forecast"     →  [-0.71, 0.33, 0.02, ...]   (far)
\`\`\`

## Properties

- **Dimension is fixed** by the model. You can't mix dimensions from different models.
- **Magnitude is not meaning.** Cosine similarity compares *direction*, not length. Always normalize or use cosine.
- **Embeddings are model-specific.** An OpenAI embedding and a Cohere embedding of the same text are not comparable.

## What they enable

- **Semantic search** — find documents by meaning, not keywords.
- **Clustering** — group similar items.
- **Classification** — train a cheap classifier on top of embeddings.
- **Deduplication** — find near-duplicates by similarity.

## Key takeaway

Embeddings are fixed-size vectors where semantic similarity becomes geometric closeness. They're the foundation of vector search, RAG, and many classification tasks. Choose one embedding model and stick with it across your system.`,
  },
  {
    id: 'les-emb-2',
    topicId: 'top-emb',
    title: 'Cosine Similarity and Distance Metrics',
    summary: 'How to compare vectors, and which metric to pick.',
    position: 2,
    durationMin: 6,
    status: 'published',
    body: `Once you have vectors, you need a way to say how "close" two are. The choice of **distance metric** matters.

## Cosine similarity

Measures the angle between two vectors, ignoring their length:

\`\`\`
cos(a, b) = (a · b) / (||a|| * ||b||)
\`\`\`

Range: [-1, 1]. 1 means same direction, 0 means orthogonal, -1 means opposite. For text embeddings, values usually sit between 0.2 and 0.9.

**This is the default for text embeddings.** Most embedding models are trained with cosine in mind.

## Dot product (inner product)

If vectors are **normalized** to length 1, dot product equals cosine similarity. Many vector indexes store normalized vectors and use dot product for speed.

## Euclidean distance (L2)

Straight-line distance between two points. Sensitive to vector magnitude. Less common for text, more common for some image embeddings.

## Which to use?

- **Text embeddings** → cosine similarity (or dot product on normalized vectors).
- **Image embeddings** (some models) → Euclidean.
- **When in doubt** → check the embedding model's documentation. It usually tells you the intended metric.

Using the wrong metric won't crash, but it will silently hurt ranking quality.

## Key takeaway

For text embeddings, use cosine similarity (equivalently, dot product on normalized vectors). Match the metric to what the embedding model was trained with — wrong metrics quietly degrade search quality.`,
  },
];

const questions_embeddings: Question[] = [
  {
    id: 'q-emb-1',
    topicId: 'top-emb',
    text: 'What does an embedding represent?',
    explanation: 'An embedding is a fixed-size vector where semantically similar content maps to nearby points in vector space.',
    options: [
      { id: 'o1', text: 'A count of words in a document', isCorrect: false },
      { id: 'o2', text: 'A vector where similar content gets similar vectors', isCorrect: true },
      { id: 'o3', text: 'A compressed image format', isCorrect: false },
      { id: 'o4', text: 'A type of database index', isCorrect: false },
    ],
  },
  {
    id: 'q-emb-2',
    topicId: 'top-emb',
    text: 'Can you compare embeddings from two different embedding models?',
    explanation: 'No — embeddings are model-specific; dimensions and geometry differ and are not comparable across models.',
    options: [
      { id: 'o1', text: 'Yes, they are interchangeable', isCorrect: false },
      { id: 'o2', text: 'No, embeddings are model-specific and not comparable across models', isCorrect: true },
      { id: 'o3', text: 'Only if both are 768 dimensions', isCorrect: false },
      { id: 'o4', text: 'Only after retraining both models together', isCorrect: false },
    ],
  },
  {
    id: 'q-emb-3',
    topicId: 'top-emb',
    text: 'What does cosine similarity measure?',
    explanation: 'Cosine similarity measures the angle between two vectors, ignoring their magnitude (length).',
    options: [
      { id: 'o1', text: 'The angle between two vectors, ignoring length', isCorrect: true },
      { id: 'o2', text: 'The straight-line distance between points', isCorrect: false },
      { id: 'o3', text: 'The number of shared dimensions', isCorrect: false },
      { id: 'o4', text: 'The sum of vector components', isCorrect: false },
    ],
  },
  {
    id: 'q-emb-4',
    topicId: 'top-emb',
    text: 'When vectors are normalized to length 1, what is dot product equivalent to?',
    explanation: 'On normalized vectors, dot product equals cosine similarity, which is why indexes store normalized vectors for fast dot product.',
    options: [
      { id: 'o1', text: 'Euclidean distance', isCorrect: false },
      { id: 'o2', text: 'Cosine similarity', isCorrect: true },
      { id: 'o3', text: 'The vector magnitude', isCorrect: false },
      { id: 'o4', text: 'Zero, always', isCorrect: false },
    ],
  },
  {
    id: 'q-emb-5',
    topicId: 'top-emb',
    text: 'Which distance metric is the default for text embeddings?',
    explanation: 'Cosine similarity (or dot product on normalized vectors) is the standard for text embeddings.',
    options: [
      { id: 'o1', text: 'Cosine similarity', isCorrect: true },
      { id: 'o2', text: 'Manhattan distance', isCorrect: false },
      { id: 'o3', text: 'Hamming distance', isCorrect: false },
      { id: 'o4', text: 'Jaccard distance', isCorrect: false },
    ],
  },
  {
    id: 'q-emb-6',
    topicId: 'top-emb',
    text: 'What training objective commonly produces text embeddings?',
    explanation: 'Contrastive learning pulls related pairs together and pushes unrelated pairs apart in vector space.',
    options: [
      { id: 'o1', text: 'Contrastive learning', isCorrect: true },
      { id: 'o2', text: 'Decision tree pruning', isCorrect: false },
      { id: 'o3', text: 'K-means clustering', isCorrect: false },
      { id: 'o4', text: 'Gradient-free evolution', isCorrect: false },
    ],
  },
];

const lessons_eval: Lesson[] = [
  {
    id: 'les-eval-1',
    topicId: 'top-eval',
    title: 'Why LLM Evaluation Is Hard',
    summary: 'Outputs are open-ended; "correct" is often subjective. You need a strategy.',
    position: 1,
    durationMin: 8,
    status: 'published',
    body: `Traditional software has tests with binary outcomes: the function returns 42 or it doesn't. LLM outputs are **open-ended** — many answers are acceptable, and "correct" depends on context, tone, and intent.

## The evaluation problem

- **Many right answers.** A summary can be phrased a hundred ways.
- **Subtle errors.** A hallucinated number is wrong even if the prose sounds great.
- **No oracle.** You don't have ground truth for every possible input.
- **Cost.** Human evaluation is slow and expensive.

## Layers of evaluation

Good teams evaluate at multiple layers:

1. **Unit tests** for deterministic logic (parsing, tool-call validation, templates).
2. **Golden datasets** — a few hundred examples with reference answers; check the model's output against them.
3. **LLM-as-judge** — use a strong model to score outputs on rubrics.
4. **Human eval** — spot-checks on a sample of production traffic.

## The trap of vibes-based eval

"It looks fine to me" doesn't scale. You will ship a change that feels better and quietly regresses on edge cases. The discipline is to **turn vibes into rubrics**: write down what "good" means, then measure it.

## Key takeaway

LLM eval is multi-layered: deterministic tests where possible, golden datasets for outputs, LLM-as-judge for scale, and human spot-checks for reality. The biggest win is moving from vibes to written rubrics.`,
  },
  {
    id: 'les-eval-2',
    topicId: 'top-eval',
    title: 'LLM-as-Judge and Rubrics',
    summary: 'Using a model to grade outputs — carefully.',
    position: 2,
    durationMin: 7,
    status: 'published',
    body: `**LLM-as-judge** uses a strong model to score the outputs of another model against a written rubric. It scales where human eval can't, but has well-known failure modes.

## The basic setup

Give the judge:
1. The user's request.
2. The model's output.
3. A **rubric** — explicit criteria (correctness, completeness, tone, safety).
4. A scoring scale (e.g. 1–5) and instructions.

\`\`\`
Prompt: "Score this answer 1–5 on factual accuracy. 5 = fully correct and cited; 1 = contains a hallucination."
\`\`\`

## Known biases

- **Position bias** — judges prefer the first of two options shown. Mitigate by swapping order and averaging.
- **Verbosity bias** — judges prefer longer answers, even when they're no more accurate. Control for length.
- **Self-preference** — a model tends to prefer outputs that look like its own. Use a different family as judge when possible.
- **Poor on subtle facts** — judges miss small numeric or naming errors. Spot-check with humans.

## Calibration

Run your judge on a small **labeled set** where you know the right answer. If the judge agrees with humans 80%+, you can trust it at scale. If not, refine the rubric or switch judges.

## Key takeaway

LLM-as-judge scales evaluation but introduces biases (position, verbosity, self-preference). Write explicit rubrics, mitigate biases by swapping order and controlling length, and calibrate the judge against a human-labeled set before trusting it.`,
  },
];

const questions_eval: Question[] = [
  {
    id: 'q-eval-1',
    topicId: 'top-eval',
    text: 'Why is LLM evaluation harder than testing traditional software?',
    explanation: 'LLM outputs are open-ended: many answers can be correct, errors are subtle, and there is no ground-truth oracle for every input.',
    options: [
      { id: 'o1', text: 'Outputs are open-ended and "correct" is often subjective', isCorrect: true },
      { id: 'o2', text: 'LLMs run too slowly to test', isCorrect: false },
      { id: 'o3', text: 'Unit tests do not exist for LLMs', isCorrect: false },
      { id: 'o4', text: 'LLMs are deterministic and never vary', isCorrect: false },
    ],
  },
  {
    id: 'q-eval-2',
    topicId: 'top-eval',
    text: 'What is "LLM-as-judge"?',
    explanation: "LLM-as-judge uses a strong model to score another model's outputs against a written rubric, scaling evaluation beyond human-only review.",
    options: [
      { id: 'o1', text: 'A model scoring another model\'s outputs against a rubric', isCorrect: true },
      { id: 'o2', text: 'A human grading a model by hand', isCorrect: false },
      { id: 'o3', text: 'A model grading its own training data', isCorrect: false },
      { id: 'o4', text: 'A type of vector index', isCorrect: false },
    ],
  },
  {
    id: 'q-eval-3',
    topicId: 'top-eval',
    text: 'What is position bias in LLM-as-judge?',
    explanation: 'Judges tend to prefer whichever option is shown first; mitigate by swapping order and averaging scores.',
    options: [
      { id: 'o1', text: 'Judges prefer the first of two options shown', isCorrect: true },
      { id: 'o2', text: 'Judges prefer shorter answers', isCorrect: false },
      { id: 'o3', text: 'Judges refuse to score', isCorrect: false },
      { id: 'o4', text: 'Judges always pick option B', isCorrect: false },
    ],
  },
  {
    id: 'q-eval-4',
    topicId: 'top-eval',
    text: 'How do you calibrate an LLM judge?',
    explanation: 'Run the judge on a small human-labeled set; agreement around 80% suggests the judge is trustworthy at scale.',
    options: [
      { id: 'o1', text: 'Run it on a labeled set and compare to human scores', isCorrect: true },
      { id: 'o2', text: 'Increase the temperature to 2.0', isCorrect: false },
      { id: 'o3', text: 'Never use a rubric', isCorrect: false },
      { id: 'o4', text: 'Compare it to a random number generator', isCorrect: false },
    ],
  },
  {
    id: 'q-eval-5',
    topicId: 'top-eval',
    text: 'What is a "golden dataset" in LLM eval?',
    explanation: 'A golden dataset is a set of examples with reference answers that you check model outputs against to detect regressions.',
    options: [
      { id: 'o1', text: 'A set of examples with reference answers to check outputs against', isCorrect: true },
      { id: 'o2', text: 'A dataset made of only correct outputs', isCorrect: false },
      { id: 'o3', text: 'The training data for the judge', isCorrect: false },
      { id: 'o4', text: 'A backup of the production database', isCorrect: false },
    ],
  },
  {
    id: 'q-eval-6',
    topicId: 'top-eval',
    text: 'What is verbosity bias?',
    explanation: 'Judges tend to prefer longer answers even when they are no more accurate than shorter ones.',
    options: [
      { id: 'o1', text: 'Judges prefer longer answers regardless of accuracy', isCorrect: true },
      { id: 'o2', text: 'Judges prefer code over prose', isCorrect: false },
      { id: 'o3', text: 'Models outputting only stop words', isCorrect: false },
      { id: 'o4', text: 'A bias toward fewer tokens to save cost', isCorrect: false },
    ],
  },
];

const lessons_deploy: Lesson[] = [
  {
    id: 'les-dep-1',
    topicId: 'top-dep',
    title: 'Running LLMs in Production: Latency and Cost',
    summary: 'The two constraints that dominate every deployment decision.',
    position: 1,
    durationMin: 9,
    status: 'published',
    body: `Two numbers dominate LLM deployments: **latency** (time to first token and tokens per second) and **cost** (per-token price × volume). Every architecture choice is a tradeoff between them.

## Where latency comes from

- **Time to first token (TTFT)** — prefill latency: processing the prompt before the first output token. Grows with prompt length.
- **Time per output token (TPOT)** — decode latency: each subsequent token. Bounded by memory bandwidth, not just compute.
- **Queueing** — when traffic exceeds capacity, requests wait.

## Where cost comes from

- **Input tokens** — you pay for the whole prompt every request, even repeated prefixes.
- **Output tokens** — usually more expensive than input.
- **Volume** — millions of requests × tokens × price adds up fast.

## Levers

1. **Prompt caching** — cache repeated prefixes (system prompts, long context). Huge cost/latency wins.
2. **Smaller models for easy tasks** — route simple queries to a cheap fast model; only send hard ones to the big one.
3. **Batch APIs** — if you don't need real-time, batch jobs are often 50% cheaper.
4. **Streaming** — stream tokens to the user so perceived latency drops even if total time is the same.
5. **Output length limits** — cap max tokens; long rambling outputs cost money.

## The golden rule

> Measure real traffic. You can't optimize what you don't profile. Log TTFT, TPOT, input/output tokens, and cost per request.

## Key takeaway

Latency and cost drive every production LLM decision. Cache prefixes, route easy tasks to small models, batch when you can, stream for perceived speed, and cap output length. Measure everything before optimizing.`,
  },
  {
    id: 'les-dep-2',
    topicId: 'top-dep',
    title: 'Observability for LLM Apps',
    summary: 'Logs, traces, and evals: seeing what your system actually does.',
    position: 2,
    durationMin: 7,
    status: 'published',
    body: `An LLM app is a distributed system with a non-deterministic component in the middle. **Observability** is how you keep it from becoming a black box.

## What to log

Every request should capture:

- **Inputs** — the full prompt, including system prompt and retrieved context.
- **Outputs** — the model's full response.
- **Metadata** — model, temperature, token counts, latency, cost.
- **Tool calls** — which tools were called, with what arguments, and what they returned.
- **Trace** — the sequence of steps in an agent loop.

## Tracing > logging

A single user request may span retrieval, multiple model calls, and several tool executions. **Tracing** ties them together under one trace ID so you can see the whole path.

\`\`\`
trace_id: abc123
├─ retrieval: 120ms, 5 chunks
├─ llm_call: 890ms, 420 tokens, $0.012
├─ tool: search_orders(user=u_123) → 3 rows
└─ llm_call: 410ms, 180 tokens, $0.005
\`\`\`

## Eval in production

Beyond logging, run **continuous evals** on a sample of production traffic:

- Did the answer match a reference? (if you have one)
- Did the LLM-judge score drop below threshold?
- Did the user send a follow-up correction (a signal of a bad answer)?

Alert on eval regressions, not just uptime.

## Key takeaway

Treat an LLM app like a distributed system. Trace entire requests end-to-end, log prompts and tool calls, and run continuous evals on production traffic so you catch regressions before users tell you.`,
  },
];

const questions_deploy: Question[] = [
  {
    id: 'q-dep-1',
    topicId: 'top-dep',
    text: 'What are the two constraints that dominate LLM deployment decisions?',
    explanation: 'Latency (TTFT and tokens per second) and cost (per-token price × volume) drive nearly every deployment tradeoff.',
    options: [
      { id: 'o1', text: 'Latency and cost', isCorrect: true },
      { id: 'o2', text: 'Color scheme and font choice', isCorrect: false },
      { id: 'o3', text: 'Disk space and RAM', isCorrect: false },
      { id: 'o4', text: 'Number of developers and office location', isCorrect: false },
    ],
  },
  {
    id: 'q-dep-2',
    topicId: 'top-dep',
    text: 'What is "time to first token" (TTFT)?',
    explanation: 'TTFT is the prefill latency: time to process the prompt before the first output token appears.',
    options: [
      { id: 'o1', text: 'The prefill latency before the first output token', isCorrect: true },
      { id: 'o2', text: 'The time to train the model', isCorrect: false },
      { id: 'o3', text: 'The time between two consecutive requests', isCorrect: false },
      { id: 'o4', text: 'The time to download the model weights', isCorrect: false },
    ],
  },
  {
    id: 'q-dep-3',
    topicId: 'top-dep',
    text: 'What is prompt caching used for?',
    explanation: 'Prompt caching stores repeated prefixes (system prompts, long context) to cut both cost and latency on repeated prefixes.',
    options: [
      { id: 'o1', text: 'Storing repeated prefixes to cut cost and latency', isCorrect: true },
      { id: 'o2', text: 'Caching the user\'s browser history', isCorrect: false },
      { id: 'o3', text: 'Saving model weights to disk', isCorrect: false },
      { id: 'o4', text: 'Compiling the prompt to bytecode', isCorrect: false },
    ],
  },
  {
    id: 'q-dep-4',
    topicId: 'top-dep',
    text: 'Why route simple queries to a smaller model?',
    explanation: 'Smaller models are cheaper and faster; routing easy tasks to them reserves the expensive model for hard tasks, cutting overall cost.',
    options: [
      { id: 'o1', text: 'Smaller models are cheaper and faster, reserving big models for hard tasks', isCorrect: true },
      { id: 'o2', text: 'Smaller models are always more accurate', isCorrect: false },
      { id: 'o3', text: 'It is required by data privacy law', isCorrect: false },
      { id: 'o4', text: 'Big models do not support simple queries', isCorrect: false },
    ],
  },
  {
    id: 'q-dep-5',
    topicId: 'top-dep',
    text: 'What does tracing tie together in an LLM app?',
    explanation: 'A single request can span retrieval, multiple LLM calls, and tool executions; a trace ID links them so you can see the whole path.',
    options: [
      { id: 'o1', text: 'The whole sequence of steps in one request under a trace ID', isCorrect: true },
      { id: 'o2', text: 'Only the final answer token', isCorrect: false },
      { id: 'o3', text: 'Just the user\'s original prompt', isCorrect: false },
      { id: 'o4', text: 'All requests from all users into one log line', isCorrect: false },
    ],
  },
  {
    id: 'q-dep-6',
    topicId: 'top-dep',
    text: 'Why run continuous evals on production traffic?',
    explanation: 'Continuous evals on a sample of real traffic catch quality regressions before users report them, since uptime alone does not measure correctness.',
    options: [
      { id: 'o1', text: 'To catch quality regressions before users report them', isCorrect: true },
      { id: 'o2', text: 'To increase API latency', isCorrect: false },
      { id: 'o3', text: 'Because it is legally required everywhere', isCorrect: false },
      { id: 'o4', text: 'To replace the need for any logging', isCorrect: false },
    ],
  },
];

export const topics: Topic[] = [
  {
    id: 'top-llms',
    slug: 'llms',
    title: 'Large Language Models',
    description: 'Tokens, context windows, sampling, and the training pipeline behind modern LLMs.',
    icon: 'BrainCircuit',
    color: 'brand',
    position: 1,
    status: 'published',
    lessons: lessons_llms,
  },
  {
    id: 'top-agents',
    slug: 'agents',
    title: 'AI Agents',
    description: 'Loops, tools, planning patterns, and keeping agents productive instead of runaway.',
    icon: 'Bot',
    color: 'accent',
    position: 2,
    status: 'published',
    lessons: lessons_agents,
  },
  {
    id: 'top-mcp',
    slug: 'mcp',
    title: 'Model Context Protocol',
    description: 'The open standard for connecting models to tools, resources, and prompts.',
    icon: 'Plug',
    color: 'warning',
    position: 3,
    status: 'published',
    lessons: lessons_mcp,
  },
  {
    id: 'top-rag',
    slug: 'rag',
    title: 'Retrieval-Augmented Generation',
    description: 'Chunking, hybrid search, and reranking for grounded answers.',
    icon: 'Search',
    color: 'brand',
    position: 4,
    status: 'published',
    lessons: lessons_rag,
  },
  {
    id: 'top-emb',
    slug: 'embeddings',
    title: 'Embeddings',
    description: 'Vectors that capture meaning, similarity metrics, and how to compare them.',
    icon: 'Boxes',
    color: 'accent',
    position: 5,
    status: 'published',
    lessons: lessons_embeddings,
  },
  {
    id: 'top-eval',
    slug: 'evaluation',
    title: 'Evaluation',
    description: 'Rubrics, LLM-as-judge, and measuring what "good" means for open-ended outputs.',
    icon: 'ClipboardCheck',
    color: 'warning',
    position: 6,
    status: 'published',
    lessons: lessons_eval,
  },
  {
    id: 'top-dep',
    slug: 'deployment',
    title: 'Deployment',
    description: 'Latency, cost, observability, and running LLM apps in production.',
    icon: 'Rocket',
    color: 'brand',
    position: 7,
    status: 'published',
    lessons: lessons_deploy,
  },
];

export const allQuestions: Question[] = [
  ...questions_llms,
  ...questions_agents,
  ...questions_mcp,
  ...questions_rag,
  ...questions_embeddings,
  ...questions_eval,
  ...questions_deploy,
];

export function getTopicBySlug(slug: string): Topic | undefined {
  return topics.find((t) => t.slug === slug);
}
export function getTopicById(id: string): Topic | undefined {
  return topics.find((t) => t.id === id);
}
export function getLessonById(id: string): { lesson: Lesson; topic: Topic } | undefined {
  for (const t of topics) {
    const lesson = t.lessons.find((l) => l.id === id);
    if (lesson) return { lesson, topic: t };
  }
  return undefined;
}
export function getQuestionsForTopic(topicId: string): Question[] {
  return allQuestions.filter((q) => q.topicId === topicId);
}

// Seed quiz history for the demo authenticated user.
export const seedQuizHistory: QuizAttempt[] = [
  {
    id: 'att-seed-1',
    topicId: 'top-llms',
    topicTitle: 'Large Language Models',
    score: 4,
    total: 5,
    createdAt: Date.now() - 1000 * 60 * 60 * 24 * 6,
    answers: [
      {
        questionId: 'q-llm-1', questionText: 'What is the core training objective of a large language model during pre-training?',
        selectedOptionId: 'o2', selectedOptionText: 'Predict the next token in a sequence',
        correctOptionId: 'o2', correctOptionText: 'Predict the next token in a sequence',
        isCorrect: true, explanation: 'LLMs are trained with next-token prediction.',
      },
      {
        questionId: 'q-llm-2', questionText: 'Roughly how many characters does one token correspond to in English text?',
        selectedOptionId: 'o2', selectedOptionText: 'About 4 characters',
        correctOptionId: 'o2', correctOptionText: 'About 4 characters',
        isCorrect: true, explanation: '1 token ≈ 4 characters.',
      },
      {
        questionId: 'q-llm-3', questionText: 'Which stage of training teaches the model to follow instructions in a chat format?',
        selectedOptionId: 'o1', selectedOptionText: 'Pre-training on web text',
        correctOptionId: 'o2', correctOptionText: 'Supervised fine-tuning (SFT)',
        isCorrect: false, explanation: 'SFT trains on instruction/response pairs.',
      },
      {
        questionId: 'q-llm-4', questionText: 'What does a temperature of 0 do during sampling?',
        selectedOptionId: 'o2', selectedOptionText: 'Always picks the highest-probability token',
        correctOptionId: 'o2', correctOptionText: 'Always picks the highest-probability token',
        isCorrect: true, explanation: 'Temperature 0 is greedy decoding.',
      },
      {
        questionId: 'q-llm-5', questionText: 'What is the "lost-in-the-middle" phenomenon?',
        selectedOptionId: 'o2', selectedOptionText: 'Models attend less reliably to information in the middle of long contexts',
        correctOptionId: 'o2', correctOptionText: 'Models attend less reliably to information in the middle of long contexts',
        isCorrect: true, explanation: 'Models attend less to the middle of long contexts.',
      },
    ],
  },
  {
    id: 'att-seed-2',
    topicId: 'top-agents',
    topicTitle: 'AI Agents',
    score: 5,
    total: 5,
    createdAt: Date.now() - 1000 * 60 * 60 * 24 * 2,
    answers: [
      {
        questionId: 'q-agt-1', questionText: 'What distinguishes an AI agent from a single chat completion?',
        selectedOptionId: 'o2', selectedOptionText: 'It runs the model in a loop with tools to decide next steps',
        correctOptionId: 'o2', correctOptionText: 'It runs the model in a loop with tools to decide next steps',
        isCorrect: true, explanation: 'An agent loops and uses tools.',
      },
      {
        questionId: 'q-agt-2', questionText: 'Who actually executes a tool call emitted by the model?',
        selectedOptionId: 'o2', selectedOptionText: 'Your application code',
        correctOptionId: 'o2', correctOptionText: 'Your application code',
        isCorrect: true, explanation: 'Your code executes tools; the model only emits calls.',
      },
      {
        questionId: 'q-agt-3', questionText: 'What is the ReAct pattern?',
        selectedOptionId: 'o1', selectedOptionText: 'Reasoning + acting in alternating steps, observing between them',
        correctOptionId: 'o1', correctOptionText: 'Reasoning + acting in alternating steps, observing between them',
        isCorrect: true, explanation: 'ReAct alternates Thought and Action.',
      },
      {
        questionId: 'q-agt-4', questionText: 'Why must you validate the arguments the model sends to a tool?',
        selectedOptionId: 'o2', selectedOptionText: 'Because the model can emit malformed or unsafe arguments',
        correctOptionId: 'o2', correctOptionText: 'Because the model can emit malformed or unsafe arguments',
        isCorrect: true, explanation: 'Validate tool arguments for safety.',
      },
      {
        questionId: 'q-agt-6', questionText: 'Which is an essential safeguard for any agent loop?',
        selectedOptionId: 'o1', selectedOptionText: 'A maximum number of steps',
        correctOptionId: 'o1', correctOptionText: 'A maximum number of steps',
        isCorrect: true, explanation: 'A step cap prevents runaway loops.',
      },
    ],
  },
  {
    id: 'att-seed-3',
    topicId: 'top-rag',
    topicTitle: 'Retrieval-Augmented Generation',
    score: 3,
    total: 5,
    createdAt: Date.now() - 1000 * 60 * 60 * 5,
    answers: [
      {
        questionId: 'q-rag-1', questionText: 'What is the core idea of retrieval-augmented generation?',
        selectedOptionId: 'o2', selectedOptionText: 'Retrieve relevant context at query time, then generate using it',
        correctOptionId: 'o2', correctOptionText: 'Retrieve relevant context at query time, then generate using it',
        isCorrect: true, explanation: 'RAG retrieves then generates.',
      },
      {
        questionId: 'q-rag-2', questionText: 'Which is a common failure mode caused by bad chunking?',
        selectedOptionId: 'o1', selectedOptionText: 'Chunks too small lose surrounding context needed to answer',
        correctOptionId: 'o1', correctOptionText: 'Chunks too small lose surrounding context needed to answer',
        isCorrect: true, explanation: 'Small chunks lose context.',
      },
      {
        questionId: 'q-rag-4', questionText: 'What does hybrid search combine?',
        selectedOptionId: 'o3', selectedOptionText: 'SQL and GraphQL',
        correctOptionId: 'o1', correctOptionText: 'BM25 keyword search and vector similarity search',
        isCorrect: false, explanation: 'Hybrid search blends BM25 and vector search.',
      },
      {
        questionId: 'q-rag-5', questionText: 'What does a cross-encoder reranker do that a bi-encoder retriever does not?',
        selectedOptionId: 'o1', selectedOptionText: 'Scores query and document together for finer relevance',
        correctOptionId: 'o1', correctOptionText: 'Scores query and document together for finer relevance',
        isCorrect: true, explanation: 'Cross-encoders score jointly.',
      },
      {
        questionId: 'q-rag-7', questionText: 'Which RAG step usually dominates answer quality?',
        selectedOptionId: 'o4', selectedOptionText: 'The order of API headers',
        correctOptionId: 'o1', correctOptionText: 'The retrieval step',
        isCorrect: false, explanation: 'Retrieval dominates RAG quality.',
      },
    ],
  },
];
