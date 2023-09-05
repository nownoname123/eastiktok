package main

import (
	"demo-tiktok/comment"
	"demo-tiktok/friendlist"
	"demo-tiktok/likelist"
	"demo-tiktok/message"
	"demo-tiktok/pushlist"
	"demo-tiktok/relation"
	"demo-tiktok/user"
	"demo-tiktok/userlogin"
	"demo-tiktok/viedo"
	"demo-tiktok/writecomment"
	"log"
	"net/http"
)

func main() {
	// 注册处理函数，将点赞请求映射到LikeHandler函数
	http.HandleFunc("/douyin/favorite/action/", likelist.LikeHandler) //赞操作
	http.HandleFunc("/douyin/favorite/list/", likelist.Favouritelist) //喜欢列表
	http.HandleFunc("/douyin/user/", user.Usermessage)                //用户信息接口
	http.HandleFunc("/douyin/user/register/", userlogin.Register)     //用户注册接口
	http.HandleFunc("/douyin/user/login/", userlogin.Userlogin)       //用户登陆接口
	http.HandleFunc("/douyin/feed/", viedo.Feed)

	http.HandleFunc("/douyin/message/action/", message.Sendmessage) //发送消息
	http.HandleFunc("/douyin/message/chat/", message.Messagelist)   //聊天记录

	http.HandleFunc("/douyin/publish/list/", pushlist.Pushlist)       //发布列表接口
	http.HandleFunc("/douyin/publish/action/", viedo.Videopublish)    //投稿接口
	http.HandleFunc("/douyin/comment/list/", comment.Showcommentlist) //评论列表接口
	http.HandleFunc("/douyin/comment/action/", writecomment.Write)    //评论操作接口
	http.HandleFunc("/douyin/relation/friend/list/", friendlist.Getfriendlist)
	http.HandleFunc("/douyin/relation/action/", relation.Action)                 //关注操作接口
	http.HandleFunc("/douyin/relation/follow/list/", relation.FollowListGet)     //关注列表接口
	http.HandleFunc("/douyin/relation/follower/list/", relation.FollowerListGet) //粉丝列表接口

	// 启动HTTP服务器，监听在本地的8080端口
	err := http.ListenAndServe("0.0.0.0:8080", nil)

	//http.ListenAndServe启动的服务器会在内部进行循环，持续监听和处理HTTP请求
	if err != nil {
		log.Println("main error：", err)
		return
	}

}
