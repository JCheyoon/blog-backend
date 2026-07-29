-- Enable Row Level Security on all public tables
-- Supabase 경고(RLS Disabled in Public)를 해결하고,
-- 각 테이블에 최소 권한 원칙을 적용합니다.

-- 1. schema_migrations: 마이그레이션 관리용 시스템 테이블
--    RLS만 활성화하고 별도 Policy를 만들지 않음 → default-deny 로 보호
ALTER TABLE public.schema_migrations ENABLE ROW LEVEL SECURITY;

-- 2. categories: 누구나 조회 가능, 쓰기/수정/삭제는 DB 소유자(백엔드)만
ALTER TABLE public.categories ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Enable read access for all users"
ON public.categories
FOR SELECT
USING (true);

-- 3. posts: 누구나 조회 가능, 쓰기/수정/삭제는 인증된 관리자만
ALTER TABLE public.posts ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Enable read access for all users"
ON public.posts
FOR SELECT
USING (true);
