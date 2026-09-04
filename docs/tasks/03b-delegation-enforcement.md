# 3b. 代理チェーンつきの認可

## 目的

Agent の実効権限を「Agent 自身の権限 ∩ 委譲元ユーザーの権限」にする。

## なぜ

交差を**付与時ではなく判定時**に取る。付与時だけだと、委譲元の権限が後で失われても
Agent は持ち続ける。期限を SQL でなく `Enforce` が時刻で落とすのと同じ理由であり、
「キャッシュはポリシーの集合を持ち、判定結果は持たない」という不変条件もここで保たれる。

親 Agent がサブ Agent を呼ぶのも同じチェーンが伸びるだけ。サブ Agent 専用の
権限機構を作らないこと —— 権限を狭める仕組みが 2 つ並ぶと、どちらが効いているか
誰にも分からなくなる。

## やること

- `auth.Request` に `OnBehalfOf []string`（`[親agent, …, 大元のユーザー]`）を足す。
- `auth.Service.Enforce` を「チェーン上の全員が許可されていること」にする。
  `Explain` も同じ規則を通す —— 答えと説明が食い違ってはいけない
  （`allows` が 1 つしかないのはそのため）。
- interceptor が `hub-on-behalf-of` を読み、`delegations` を引いてチェーンを
  `contextx` に載せる。`hub-org` と同じ形（`requestedOrg` が雛形）。
- チェーンの深さを `delegations.max_depth` で切る。
- `audit_logs.channel` の CHECK に `'agent'` を足し、チェーンを記録する。
  承認者が「誰の代理で動いた Agent の行為か」を分かった上で判断できるように。
- **`escalation` は主体の種類に関係なく効かせる。** どれだけ強い委譲を受けていても
  `AddPermissionsToRole` / `AddRolesToGroup` / `AddGroupsToUser` /
  `DecideAccessRequest` には触れない。委譲は escalation を通過する理由にならない。

## 完了条件

- Agent が委譲元より広い権限を行使できない（委譲元の権限を落とすと即座に狭まる）。
- チェーンが深くなるほど権限が広がることはない（単調減少のテスト）。
- 期限切れ・失効済みの委譲では代理を名乗れない。
- 監査ログからチェーンを辿れる。
