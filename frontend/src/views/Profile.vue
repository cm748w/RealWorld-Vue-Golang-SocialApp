<template>
	<q-page class="constrain q-pa-md">
		<div class="row q-col-gutter-lg constrain">
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
			<div class="col-12">
				<q-separator inset />
			</div>
			<div class="col-4" v-for="post in userPosts" :key="post._id">
				<Post :post="post" />
			</div>
		</div>

	</q-page>





</template>

<script>
// @ is an alias to /src
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
		// 获取所有用户信息和帖子
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