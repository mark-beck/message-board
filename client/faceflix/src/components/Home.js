import React, { useState, useEffect } from "react";
import { MDBRow, MDBCard, MDBCardBody, MDBIcon } from "mdb-react-ui-kit";
import Button from "react-bootstrap/Button";

import UserService from "../services/user.service";

import "./Home.css"
import Post from "./Post";
import PostModal from "./PostModal"

const Home = () => {
    const [content, setContent] = useState("");
    const [createshow, setCreateShow] = useState(false);

    const reload_posts = () => {
        console.log("reloading post")
        UserService.getPublicContent().then((posts) => {
            console.log(posts)
            let c = posts.map((post) => {
                console.log("posts:" + post)
                return (
                    <Post postData={post} />
                );
            })
            setContent(c)
        }).catch((error) => {
            console.log("posts error: " + error)
        })
    }

    useEffect(() => {
        UserService.getPublicContent().then((posts) => {
            console.log(posts)
            let c = posts.map((post) => {
                console.log(post)
                return (
                    <Post postData={post} />
                );
            })
            setContent(c)
        }).catch((error) => {
            console.log(error)
        })
    }, []);

    const handleClose = () => setCreateShow(false);
    const handleShow = () => setCreateShow(true);

    return (
        <div className="container">
            <header className="jumbotron"></header>
            <Button variant="primary" onClick={handleShow}>
                create
            </Button>
            <PostModal handleClose={handleClose} show={createshow} reload_posts={reload_posts} />


            {content}

        </div>

    );
};

export default Home;
