-- 000005 롤백: posts 조회 정책을 전체 공개로 되돌립니다.

DROP POLICY IF EXISTS "Enable read published posts for all users" ON public.posts;

CREATE POLICY "Enable read access for all users"
ON public.posts
FOR SELECT
USING (true);
