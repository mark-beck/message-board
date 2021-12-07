import React, { useState, useEffect } from "react";

import UserService from "../services/user.service";
import UserList from "./userList/UserList";
import Button from "react-bootstrap/Button";
import CreateForm from "./userList/CreateForm";
import { propTypes } from "react-bootstrap/esm/Image";

const BoardUser = (props) => {
    const [content, setContent] = useState("");
    const [createshow, setCreateShow] = useState(false);

    const reload_list = () => {
        console.log("reloading")
        UserService.getUsers().then(
            (response) => {
                setContent(<UserList userlist={response.data} reload_list={reload_list} addToast={props.addToast} />);
            },
            (error) => {
                const _content =
                    (error.response &&
                        error.response.data &&
                        error.response.data.message) ||
                    error.message ||
                    error.toString();

                setContent(_content);
            }
        );
    }

    useEffect(() => {
        UserService.getUsers().then(
            (response) => {
                setContent(<UserList userlist={response.data} reload_list={reload_list} addToast={props.addToast} />);
            },
            (error) => {
                const _content =
                    (error.response &&
                        error.response.data &&
                        error.response.data.message) ||
                    error.message ||
                    error.toString();

                setContent(_content);
            }
        );
    }, []);

    const handleClose = () => setCreateShow(false);
    const handleShow = () => setCreateShow(true);

    return (
        <div className="container">
            <header className="jumbotron">
                <Button variant="primary" onClick={handleShow}>
                    create
                </Button>
                <CreateForm handleClose={handleClose} show={createshow} reload_list={reload_list} />
                {content}
            </header>
        </div>
    );
};

export default BoardUser;
