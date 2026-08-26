<template>
	<q-page class="constrain home-page">
		<div class="row q-col-gutter-lg">
			<div class="col-md-3 gt-sm">
				<SideBar />
			</div>

			<div class="col-12 col-md-6 q-mx-auto">
				<div class="q-pa-md">

					<transition name="fade" mode="out-in">
						<!-- Skeletons -->
						<div v-if="!load" key="skeleton">
							<div v-for="i in 3" :key="i" class="surface q-mb-md q-pa-md skeleton-card">
								<div class="row items-center">
									<q-skeleton type="QAvatar" size="48px" animation="wave" />
									<div class="q-ml-md col">
										<q-skeleton type="text" animation="wave" style="width: 40%" />
										<q-skeleton type="text" animation="wave" style="width: 25%" />
									</div>
								</div>
								<q-skeleton height="200px" square animation="wave" class="q-mt-md rounded-borders" />
								<div class="row q-mt-md q-gutter-sm">
									<q-skeleton type="QBtn" animation="wave" />
									<q-skeleton type="QBtn" animation="wave" />
									<q-skeleton type="QBtn" animation="wave" />
								</div>
							</div>
						</div>

						<!-- Posts -->
						<div v-else key="posts">
							<q-infinite-scroll @load="onLoad" :offset="250">
								<Post v-for="post in posts" :key="post._id" :post="post" />

								<template v-slot:loading>
									<div class="q-pa-lg text-center">
										<q-spinner color="" size="3em" />
										<div class="q-mt-md muted">Loading more posts...</div>
									</div>
								</template>

								<template v-slot:end v-if="posts.length > 0">
									<div class="q-pa-md text-center muted">
										<q-icon name="eva-inbox-outline" size="24px" />
										<div class="q-mt-sm">No More Posts</div>
									</div>
								</template>
							</q-infinite-scroll>
						</div>
					</transition>

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
			this.max = 0
			this.posts = []
			this.hasReachedEnd = false
			this.load = false
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
		async onLoad(index, done) {
			if (this.hasReachedEnd) {
				done(true)
				return
			}
			try {
				if (this.current < this.max) {
					this.current++
					await this.GetAllPosts(true)
				}
				done(this.hasReachedEnd)
			} catch (error) {
				console.error('error loading more posts', error)
				done(true)
			}
		},
	},
	async mounted() {
		await this.GetAllPosts()
	},
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
	transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
	opacity: 0;
}

.skeleton-card {
	animation: skeletonPulse 1.4s ease-in-out infinite;
}

@keyframes skeletonPulse {
	0%, 100% {
		opacity: 0.55;
	}
	50% {
		opacity: 1;
	}
}
</style>
