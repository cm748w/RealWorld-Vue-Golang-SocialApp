describe("Auth Component Tests", () => {
    beforeEach(() => {
        cy.visit('/Auth')
    })

    
    describe('Page Layout', () => {
        it('should display signin and signup cards', () => {
            // check for bot cards exists
            cy.get('.q-card').should('have.length', 2)

            // check titles
            cy.contains('登录').should('be.visible')
            cy.contains('注册').should('be.visible')
        })

        it('should have proper layout structure', () => {
            cy.get('.col-5').should('exist')
            cy.get('.col-7').should('exist')
            cy.get('.row').should('exist')

        })
    })

    describe('登录表单', ()=>{
        it('should display all signin form elements', ()=>{
            cy.get('.col-5 input').should('have.length', 2)

            // check for buttom
            cy.contains('登录').should('be.visible')
            cy.get('.col-5 button[type="submit"]').should('be.visible')
        })

        it('should allow typing in signin inputs', ()=> {
            cy.get('.col-5 input').eq(0)
                .type('test@example.com')
                .should('have.value', 'test@example.com')

            cy.get('.col-5 input').eq(1)
                .type('password123')
                .should('have.value', 'password123')
        })

        it('should have password input type', ()=> {
            cy.get('.col-5 input').eq(1)
                .should('have.attr', 'type', 'password')
        })
    })

    describe('注册表单', ()=> {
        it('should display all signup form elements', ()=> {
            // 检查所有的输入是否存在
            cy.get('.col-7 input').should('have.length', 4)
            cy.contains('名字 *').should('be.visible')
            cy.contains('姓氏 *').should('be.visible')
            cy.contains('邮箱 *').should('be.visible')
            cy.contains('密码 *').should('be.visible')

            // 检查按钮
            cy.contains('注册').should('be.visible')
        })

        it('should allow typing in all signup inputs', ()=> {
            cy.get('.col-7 input').eq(0)
                .type('John')
                .should('have.value', 'John')

            cy.get('.col-7 input').eq(1)
                .type('Doe')
                .should('have.value', 'Doe')

            cy.get('.col-7 input').eq(2)
                .type('j@example.com')
                .should('have.value', 'j@example.com')

            cy.get('.col-7 input').eq(3)
                .type('password123')
                .should('have.value', 'password123')
        })

        it('should have corrent buttom colors', ()=> {
            // 登录
            cy.get('.col-5 .q-btn').should('have.class', 'bg-primary')

            // 注册
            cy.get('.col-7 .q-btn').should('have.class', 'bg-positive')

        })
    })

    describe('登录表单互动', ()=> {
        it('处理空输入请求', ()=> {
            cy.get('.col-5 button[type="submit"]').click()
            cy.get('.q-notification').should('be.visible').and('contain', '请输入邮箱')
        })
    })
    describe('注册表单互动', ()=> {
        it('处理空输入请求', ()=> {
            cy.get('.col-7 button[type="submit"]').click()
            cy.get('.q-notification').should('be.visible').and('contain', '请输入姓名')
        })
    })

    describe('响应式设计', ()=> {
        it('多端下维持布局稳定', ()=> {
            cy.viewport(375, 667)
            cy.get('.q-card').should('be.visible')

            cy.viewport(768, 1024)
            cy.get('.col-5').should('be.visible')
            cy.get('.col-7').should('be.visible')

            cy.viewport(1200, 800)
            cy.get('.row').should('be.visible')
        })
    })
})

