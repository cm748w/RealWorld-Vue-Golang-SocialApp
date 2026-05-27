<template>
	<q-page class="constrain q-pa-md home-page">
		<div class="row q-col-gutter-lg">
			<div class="col-3">
				<SideBar />
			</div>
			<div v-if="!load" class="col-6 q-mx-auto">
				<div class="q-pa-md">
					<q-card>
						<q-item>
							<q-item-section avatar>
								<q-skeleton type="QAvatar" />
							</q-item-section>

							<q-item-section>
								<q-item-label>
									<q-skeleton type="text" />
								</q-item-label>
								<q-item-label caption>
									<q-skeleton type="text" />
								</q-item-label>
							</q-item-section>
						</q-item>

						<q-skeleton height="200px" square />
						<q-card-actions class="q-gutter-md">
							<q-skeleton type="QBtn" />
							<q-skeleton type="QBtn" />
						</q-card-actions>
					</q-card>
				</div>
			</div>
			<div v-else class="col-6 q-mx-auto">
				<Post v-for="post in posts" :key="post._id" :post="post" />

				<!-- loading indicatior for more posts -->
				<div v-if="loadingMore" class="q-pa-lg text-center">
					<q-spinner-hourlass color="primary" size="3em" />
					<div class="q-mt-md text-grey-7">
						Loading more posts...
					</div>
				</div>

				<!-- end of posts indicatior -->
				<div v-if="hasReachedEnd && posts.length > 0" class="q-pa-md text-center text-grey-6">
					<q-icon name="eva-inbox-outline" size="24px" />
					<div class="q-mt-sm">No More Posts</div>
				</div>

				<div class="bottom-spacer"></div>

			</div>
			<div class="col-3">
				<Rightbar />
			</div>
			<Add @created="OnPostCreated" />
		</div>
		<div class="q-pa-lg flex justify-center pagination-row">
			
		</div>
	</q-page>
</template>

<script>
import Add from '@/components/post/Add.vue'
import Post from '@/components/post/Post.vue'
import SideBar from '@/components/sideBar/SideBar.vue'
import Rightbar from '@/components/rightbar/Rightbar.vue'
import { mapActions } from 'vuex'
export default {
	name: 'HomeView',
	data() {
		return {
			current: 1,
			max: 0,
			posts: [],
			load: false,
			hasReachedEnd: false,
		}
	},
	// watch: {
	// 	current() {
	// 		this.GetAllPosts()
	// 	}
	// },
	components: {
		Add,
		Post,
		SideBar,
		Rightbar,
	},
	methods: {
		...mapActions({ getPosts: 'GetAllPosts' }),
		async OnPostCreated() {
			this.current = 1
			await this.GetAllPosts()
		},
		async GetAllPosts(append = false) {
			try {
				const data = await this.getPosts(this.current)
				console.log("post data", data)
				if (data?.data) {
					this.max = data?.numberOfPages
					console.log("Is append", append)
					if (append) {
						this.posts = [...this.posts, ...data.data]
					} else {
						this.posts = data?.data
					}

					// check if we reached the end
					this.hasReachedEnd = this.current >= this.max
				}

				if (data) {
					this.load = true
				}
			} catch (error) {
				console.error("Error loading posts", error)
			}
		},
		async loadMorePosts(){
			if (this.loadingMore || this.hasReachedEnd) {
				return
			}

			if (this.current < this.max) {
				this.loadingMore = true
				this.current++

				try {
					await this.GetAllPosts(true)
					await new Promise(resolve => setTimeout(resolve, 1000))
				} catch (error) {
					console.error('error loading more posts', error)
					this.current--
				} finally {
					this.loadingMore = false
				}
			}
		},
		handleScroll(){
			const scrollTop = window.pageYOffset || document.documentElement.scrollTop
			const windowHeight = window.innerHeight
			const documentHeight = document.documentElement.scrollHeight

			if (scrollTop + windowHeight >= documentHeight - 200) {
				this.loadMorePosts()
			}
		},
	},
	async mounted() {
		
		setTimeout(async () => {
			await this.GetAllPosts()

			window.addEventListener('scroll', this.handleScroll)
		}, 1000)
	},
	beforeUnmount(){
		window.removeEventListener('scroll', this.handlwScroll)
	}
}
</script>

<style scoped>
.pagination-row {
	margin-top: 24px;
}
</style>
