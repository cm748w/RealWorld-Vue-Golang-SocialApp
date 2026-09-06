// 复现用户报告的场景：已登录用户访问他人 /profile/:id 后点击 FOLLOW。
// 回归：ShowProfile.vue 的 created() 被误放进 methods，debouncedFollow 从未初始化，
// 点击 Follow/Following 静默无反应（commit 74711e1 引入）。
// 断言的关键是：点击后必须真的发出 PATCH /api/user/:id/following 请求，
// 按钮文案 Follow <-> Following 翻转，且服务端 followers 数组真实变化。
describe("Follow / Unfollow from another user's profile", () => {
    const stamp = Date.now();
    const emailA = `actor_${stamp}@cy.test`;
    const emailB = `target_${stamp}@cy.test`;
    const password = 'Password123!';

    function signup(email, firstName, lastName) {
        return cy.request({
            method: 'POST',
            url: '/api/user/signup',
            body: { email, password, firstName, lastName },
            failOnStatusCode: false,
        }).then((res) => {
            expect(res.status, 'signup 201').to.eq(201);
            expect(res.body.result).to.have.property('_id');
            expect(res.body.token).to.be.a('string').and.not.empty;
            return res.body; // { result, token }
        });
    }

    function fetchUser(id, token) {
        return cy.request({
            method: 'GET',
            url: `/api/user/getUser/${id}`,
            headers: { Authorization: `Bearer ${token}` },
            failOnStatusCode: false,
        }).then((res) => {
            expect(res.status, 'getUser 200').to.eq(200);
            const u = res.body.user || res.body.result || res.body.data || res.body;
            expect(u).to.have.property('_id');
            return u;
        });
    }

    it('点击 Follow 关注他人，再点击取消关注', () => {
        // 捕获点击 Follow/取消时必须发出的写请求；若 handler 仍是死链则此请求永不发出，测试超时失败
        cy.intercept('PATCH', '/api/user/*/following').as('followReq');

        let actor = null;
        let target = null;

        signup(emailB, 'Target', 'User').then((b) => { target = b; return signup(emailA, 'Actor', 'User'); })
            .then((a) => { actor = a; });

        cy.then(() => {
            expect(actor, 'actor signed up').to.not.be.null;
            expect(target, 'target signed up').to.not.be.null;

            // 进入 app 域后写入登录态，再访问他人主页（应用基路径为 /app）
            cy.visit('/app/Auth');
            cy.window().then((win) => {
                win.localStorage.setItem('profile', JSON.stringify({ result: actor.result, token: actor.token }));
            });
            cy.visit(`/app/profile/${target.result._id}`);
        });

        // 页面渲染出 Follow 按钮（v-else-if 分支，非本人）
        cy.contains('button', /^Follow$/, { timeout: 20000 }).should('be.visible');

        // ---- 第一次点击：关注 ----
        cy.contains('button', /^Follow$/).click();

        // 关键断言：真实发出了关注请求（修复前这里会直接超时 = 复现 bug）
        cy.wait('@followReq', { timeout: 15000 }).its('response.statusCode').should('eq', 200);

        // 按钮应翻转为 Following 且出现成功提示
        cy.contains('button', /^Following$/, { timeout: 20000 }).should('be.visible');
        cy.get('.q-notification', { timeout: 5000 }).should('contain', '关注成功');

        // 服务端校验：目标用户的 followers 已包含当前登录用户
        cy.then(() => fetchUser(target.result._id, actor.token)).then((u) => {
            const followers = (u.followers || []).map((x) => String(x));
            expect(followers, `followers=${JSON.stringify(followers)}`).to.include(
                String(actor.result._id), 'actor 应出现在 target.followers'
            );
        });

        // 防抖窗口为 800ms（leading 模式会丢弃窗口内的重复点击），切换前先等窗口过去
        cy.wait(1200);

        // ---- 第二次点击：取消关注 ----
        cy.contains('button', /^Following$/).click();
        cy.wait('@followReq', { timeout: 15000 }).its('response.statusCode').should('eq', 200);

        cy.contains('button', /^Follow$/, { timeout: 20000 }).should('be.visible');
        cy.get('.q-notification', { timeout: 5000 }).should('contain', '取消关注成功');

        cy.wait(1200);

        cy.then(() => fetchUser(target.result._id, actor.token)).then((u) => {
            const followers = (u.followers || []).map((x) => String(x));
            expect(followers).not.to.include(String(actor.result._id), 'actor 应从 target.followers 移除');
        });
    });
});
