-- Private (draft) posts must not be publicly readable.
-- Supabase REST API(anon key)는 이 policy를 기준으로 조회하므로,
-- published = true 인 행만 공개하고 초안은 default-deny로 보호합니다.
-- (Go 백엔드는 postgres 소유자 역할로 접속해 RLS를 우회하므로
--  관리자 CRUD에는 영향이 없습니다.)

DROP POLICY IF EXISTS "Enable read access for all users" ON public.posts;

CREATE POLICY "Enable read published posts for all users"
ON public.posts
FOR SELECT
USING (published = true);
