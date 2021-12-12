import React, { useState, useEffect } from "react";

import UserService from "../services/user.service";

import "./Home.css"
import Post from "./Post";

const Home = () => {
    const [content, setContent] = useState("");

    useEffect(() => {
        let c = UserService.getPublicContent().then((posts) => {
            console.log(posts)
            posts.map((post) => {
                console.log(post)
                return (
                    <Post author={post.author} text={post.text} />
                );
            })
            setContent(c)
        }).catch((error) => {
            console.log(error)
        })
    }, []);

    return (
        <div className="container">
            <header className="jumbotron">
                <div>{content}</div>
            </header>

        </div>
    );
};

export default Home;
