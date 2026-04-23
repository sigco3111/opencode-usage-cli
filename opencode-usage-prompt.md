# OpenCode 사용량 조회 CLI 앱 개발 프롬프트

## 목적

`~/.local/share/opencode/opencode.db` (SQLite) 에서 사용량 데이터를 조회해서 터미널에 예쁘게 출력하는 CLI 도구.

## 데이터베이스 구조

### 메시지 테이블 스키마

```sql
CREATE TABLE message (
  id text PRIMARY KEY,
  session_id text NOT NULL,
  time_created integer NOT NULL,  -- Unix ms
  time_updated integer NOT NULL,
  data text NOT NULL  -- JSON
);
```

### message.data JSON 구조 (assistant 메시지)

```json
{
  "role": "assistant",
  "time": { "created": 1772684408629, "completed": 1772684657425 },
  "modelID": "glm-5.1",
  "providerID": "zai-coding-plan",
  "agent": "build",
  "cost": 0,
  "tokens": {
    "input": 12345,
    "output": 6789,
    "reasoning": 100,
    "cache": { "read": 0, "write": 0 }
  }
}
```

### session 테이블 스키마

```sql
CREATE TABLE session (
  id text PRIMARY KEY,
  project_id text NOT NULL,
  directory text NOT NULL,
  title text NOT NULL,
  time_created integer NOT NULL,
  time_updated integer NOT NULL
);
```

## 핵심 SQL 쿼리

### 1. 기간별 전체 요약

```sql
SELECT
  COUNT(*) as total_assistant_messages,
  SUM(json_extract(data, '$.tokens.input')) as total_input_tokens,
  SUM(json_extract(data, '$.tokens.output')) as total_output_tokens,
  SUM(json_extract(data, '$.tokens.reasoning')) as total_reasoning_tokens,
  SUM(json_extract(data, '$.cost')) as total_cost
FROM message
WHERE time_created >= {start_ms}
  AND json_extract(data, '$.role') = 'assistant';
```

### 2. 유저 요청 수

```sql
SELECT COUNT(*) FROM message
WHERE time_created >= {start_ms}
  AND json_extract(data, '$.role') = 'user';
```

### 3. 모델별 사용량

```sql
SELECT
  json_extract(data, '$.modelID') as model,
  json_extract(data, '$.providerID') as provider,
  COUNT(*) as messages,
  SUM(json_extract(data, '$.tokens.input')) as input_tokens,
  SUM(json_extract(data, '$.tokens.output')) as output_tokens,
  SUM(json_extract(data, '$.cost')) as total_cost
FROM message
WHERE time_created >= {start_ms}
  AND json_extract(data, '$.role') = 'assistant'
GROUP BY model, provider ORDER BY messages DESC;
```

### 4. 날짜별 사용량

```sql
SELECT
  date(time_created / 1000, 'unixepoch', '+9 hours') as date,
  COUNT(*) as messages,
  SUM(json_extract(data, '$.tokens.input')) as input_tokens,
  SUM(json_extract(data, '$.tokens.output')) as output_tokens,
  SUM(json_extract(data, '$.cost')) as daily_cost
FROM message
WHERE time_created >= {start_ms}
  AND json_extract(data, '$.role') = 'assistant'
GROUP BY date ORDER BY date;
```

### 5. 프로젝트별 세션 수

```sql
SELECT
  s.directory as project,
  COUNT(*) as sessions,
  MIN(date(s.time_created / 1000, 'unixepoch', '+9 hours')) as first_used,
  MAX(date(s.time_updated / 1000, 'unixepoch', '+9 hours')) as last_used
FROM session s
WHERE s.time_created >= {start_ms}
GROUP BY s.directory ORDER BY sessions DESC;
```

### 6. 시간별 사용량 (피크 시간대 탐지)

```sql
SELECT
  strftime('%Y-%m-%d %H:00', datetime(time_created / 1000, 'unixepoch', '+9 hours')) as hour,
  COUNT(*) as messages,
  SUM(json_extract(data, '$.tokens.input')) as input_tokens,
  SUM(json_extract(data, '$.tokens.output')) as output_tokens
FROM message
WHERE time_created >= {start_ms}
  AND json_extract(data, '$.role') = 'assistant'
GROUP BY hour
ORDER BY messages DESC;
```

## 요구사항

1. **언어/프레임워크**: Go 또는 Python 중 프로젝트 구조에 맞는 것 선택. (빠른 바이너리 배포 가능하면 Go 우선)
2. **CLI 옵션**:
   - `--period today|week|month|all` (기본값: month)
   - `--from YYYY-MM-DD --to YYYY-MM-DD` (커스텀 기간)
   - `--by-model` 모델별 상세 보기
   - `--by-day` 일별 트렌드 보기
   - `--by-project` 프로젝트별 보기
   - `--by-hour` 시간별 피크 분석 보기
   - DB 경로 오버라이드 `--db-path`
3. **출력 포맷**: 터미널 테이블 (aligned columns). 토큰 수는 human-readable (e.g. 1.6억, 1574만, 281K 등 한국어 단위 사용)
4. **타임존**: KST (UTC+9) 기준. SQL에서 `'+9 hours'` 적용
5. **DB 경로 기본값**: `~/.local/share/opencode/opencode.db`
6. **에러 처리**: DB 파일 없을 때, 테이블 없을 때 명확한 에러 메시지
7. **저장 위치**: 프로젝트 루트에 `opencode-usage/` 디렉토리 생성해서 그 안에
8. **바이너리 이름**: `oc-usage`

## 출력 예시

기본 실행(`oc-usage` 또는 `oc-usage --period month`) 시 아래 **4개 섹션을 모두** 출력한다:

### 섹션 1: 전체 요약

```
📊 OpenCode 사용량 (2026.03.23 ~ 2026.04.23)

전체 요약
─────────────────────────────────────────
  요청 수        3,552회
  응답 수        25,334회
  Input Tokens   1.61억
  Output Tokens  1,574만
  Reasoning      281만
  총 비용        $0
─────────────────────────────────────────
```

### 섹션 2: 모델별 사용량 TOP 5

```
모델별 사용량 TOP 5
─────────────────────────────────────────────────────────────────────
  #  모델           Provider          응답수    Input     Output
  1  glm-5.1        zai-coding-plan   15,917    1.09억    1,045만
  2  glm-5          zai-coding-plan    3,826    2,084만    250만
  3  gpt-5-mini     github-copilot     1,829      894만    104만
  4  glm-4.5        zai-coding-plan    1,358    1,004만     64만
  5  glm-4.6v       zai-coding-plan    1,052      295만     44만
─────────────────────────────────────────────────────────────────────
```

### 섹션 3: 프로젝트별 세션 수 TOP 5

```
프로젝트별 세션 수 TOP 5
──────────────────────────────────────────────────────────────
  #  프로젝트                        세션수   기간
  1  bookk                             854    03/30 ~ 04/23
  2  my-startup                        261    04/20 ~ 04/21
  3  ubob2025                          244    03/05 ~ 04/14
  4  ubob2030                           53    04/09 ~ 04/22
  5  ubob2020                           43    03/09 ~ 03/24
──────────────────────────────────────────────────────────────
```

### 섹션 4: 일별 사용 추이 + 피크 시간

```
일별 사용 추이
────────────────────────────────────────────────
  날짜        응답수    Input      Output
  03/23          90      62만        5.6만
  03/25         129      80만       16.3만
  03/26         441     265만       47.3만
  ...
  04/20       2,228   1,108만       99.7만
  04/21       3,135   1,394만      117.2만  ⬆ 최대
  04/22       1,535     574만       60.7만
  04/23       1,032     343만       49.2만
────────────────────────────────────────────────

🔥 한 달 최고 피크: 2026-04-21 (월)
   하루 응답 3,135회 / Input 1,394만 / Output 117.2만

   ⏰ 최고 피크 시간대: 2026-04-21 07:00 ~ 08:00
      1시간 응답 451회 / Input 162.7만
```

## 플래그별 상세 모드

- `--by-model`  실행 시: TOP 5가 아닌 **전체 모델** 목록 출력
- `--by-day`    실행 시: 일별 표에 **누적 합계** 행 추가
- `--by-project` 실행 시: TOP 5가 아닌 **전체 프로젝트** 목록 출력
- `--by-hour`   실행 시: **전체 시간별 데이터** 출력 (요청수 기준 내림차순)

플래그 없이 실행하면 기본 4섹션 요약 화면.
