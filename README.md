# oc-usage

OpenCode 사용량 통계를 터미널에서 조회하는 CLI 도구.

`~/.local/share/opencode/opencode.db` (SQLite)에서 데이터를 읽어 한국어 포맷팅된 테이블로 출력합니다.

## 설치

```bash
cd opencode-usage && make install
```

또는 직접 빌드:

```bash
cd opencode-usage && CGO_ENABLED=0 go build -o oc-usage .
```

## 사용법

### 기본 실행 (최근 30일)

```bash
oc-usage
```

4개 섹션이 출력됩니다: 전체 요약, 모델별 TOP 5, 프로젝트별 TOP 5, 일별 추이 + 피크

### 기간 선택

```bash
oc-usage --period today
oc-usage --period week
oc-usage --period month    # 기본값
oc-usage --period all
oc-usage --from 2026-04-01 --to 2026-04-15
```

### 상세 보기

```bash
oc-usage --by-model      # 전체 모델 목록
oc-usage --by-day         # 일별 추이 + 누적 합계
oc-usage --by-project     # 전체 프로젝트 목록
oc-usage --by-hour        # 시간별 분포
oc-usage --by-agent       # 에이전트별 사용량
```

### JSON 출력

```bash
oc-usage --json
oc-usage --json --period week
```

### 기타 옵션

```bash
oc-usage --version
oc-usage --db-path /path/to/custom.db
oc-usage --color=never
oc-usage --color=always
```

## 요구사항

- Go 1.21+
- OpenCode가 설치되어 있고 `~/.local/share/opencode/opencode.db` 가 존재해야 함

## License

MIT
