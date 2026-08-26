<template>
	<q-page class="constrain q-pa-md">
		<ShowProfile
			:userData="userData"
			:userPosts="userPosts"
			:isSameUser="isSameUser"
			@EditProfile="EditMode = !EditMode"
			@update-user="updateUserLocal"
			v-if="!EditMode" />

		<EditProfile
			:userData="userData"
			:isSameUser="isSameUser"
			@EditProfile="EditMode = !EditMode"
			@update-user="updateUserLocal"
			v-else />

		<div class="row q-col-gutter-md q-pt-md">
			<div class="col-12">
				<div class="section-title q-pa-md q-pb-xs">Posts</div>
			</div>

			<div class="col-12 col-sm-6 col-md-4" v-for="post in userPosts" :key="post._id">
				<Post :post="post" />
			</div>

			<div v-if="!userPosts.length" class="col-12 text-center muted q-pa-lg">
				No posts yet.
			</div>
		</div>
	</q-page>
</template>

<script>
import { mapGetters, mapMutations, mapActions } from 'vuex'
import Post from '@/components/post/Post.vue'
import ShowProfile from '@/components/user/ShowProfile.vue'
import EditProfile from '@/components/user/EditProfile.vue'

export default {
	name: 'ProfileView',
	data() {
		return {
			userPosts: [],
			EditMode: false,
		}
	},
	watch: {
		$route() {
			this.EditMode = false
			this.GetAll()
		}
	},
	mounted() {
		console.log("userid:", this.$route.params.id)
		this.SetData()
		this.GetAll()
	},
	computed: {
		...mapGetters("auth", ['GetUserData']),
		...mapGetters("users", {
			GetCachedUser: 'GetUser'
		}),
		userData() {
			return this.GetCachedUser(this.$route.params.id) || {}
		},
		isSameUser() {
			return String(this.GetUserData?.result?._id) == String(this.$route.params.id)
		}
	},
	methods: {
		...mapMutations("auth", ['SetData']),
		...mapMutations("users", {
			SetCachedUser: 'SetUser'
		}),
		...mapActions("users", ['GetUserById']),
		updateUserLocal(updatedUser) {
			if (updatedUser?._id) {
				this.SetCachedUser(updatedUser)
			}
		},
		async GetAll() {
			const profileid = this.$route.params.id

			const data = await this.GetUserById(profileid)
			this.userPosts = data?.posts || []
		}
	},
	components: {
		ShowProfile,
		EditProfile,
		Post,
	},
}
</script>
