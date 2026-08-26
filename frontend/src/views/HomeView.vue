<template>
	<q-page class="constrain home-page">
		<div class="row q-col-gutter-lg">
			<div class="col-md-3 gt-sm">
				<SideBar />
			</div>

			<div class="col-12 col-md-6 q-mx-auto">
				<div class="q-pa-md">

					<!-- Skeletons -->
					<template v-if="!load">
						<div v-for="i in 3" :key="i" class="surface q-mb-md q-pa-md">
							<div class="row items-center">
								<q-skeleton type="QAvatar" size="48px" />
								<div class="q-ml-md col">
									<q-skeleton type="text" style="width: 40%" />
									<q-skeleton type="text" style="width: 25%" />
								</div>
							</div>
							<q-skeleton height="200px" square class="q-mt-md rounded-borders" />
							<div class="row q-mt-md q-gutter-sm">
								<q-skeleton type="QBtn" />
								<q-skeleton type="QBtn" />
								<q-skeleton type="QBtn" />
							</div>
						</div>
					</template>

					<!-- Posts -->
					<template v-else>
						<Post v-for="post in posts" :key="post._id" :post="post" />

						<div v-if="loadingMore" class="q-pa-lg text-center">
							<q-spinner color="primary" size="3em" />
							<div class="q-mt-md muted">Loading more posts...</div>
						</div>

						<div v-if="hasReachedEnd && posts.length > 0" class="q-pa-md text-center muted">
							<q-icon name="eva-inbox-outline" size="24px" />
							<div class="q-mt-sm">No More Posts</div>
						</div>

						<div class="bottom-spacer" />
					</template>

				</div>
			</div>

			<div class="col-md-3 gt-sm">
				<Rightbar />
			</div>

			<Add @created="OnPostCreated" />
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
		window.removeEventListener('scroll', this.handleScroll)
	}
}
</script>
