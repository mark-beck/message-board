import React, { useState, useEffect } from "react";
import { Switch, Route, Link } from "react-router-dom";
import { LinkContainer } from 'react-router-bootstrap';
import "bootstrap/dist/css/bootstrap.min.css";
import "./App.css";

import AuthService from "./services/auth.service";

import Login from "./components/Login";
import Home from "./components/Home";
import BoardUser from "./components/BoardUser";

import { Navbar, Nav, Toast, ToastContainer } from "react-bootstrap";

const App = () => {

    const [currentUser, setCurrentUser] = useState(undefined);
    const [toastQueue, setToastQueue] = useState([])

    useEffect(() => {
        const user = AuthService.getCurrentUser();

        if (user) {
            setCurrentUser(user);
        }
    }, []);

    const logOut = () => {
        AuthService.logout();
    };

    const addToast = (toast) => {
        let tmp = toastQueue.slice();
        tmp.push(toast);
        setToastQueue(tmp);
        console.log("toastqueue: " + toastQueue);
    }

    const renderToasts = () => {
        console.log("render toast: toastqueue: " + toastQueue);
        if (toastQueue) {
            return toastQueue.map((toast, i) => {
                return (
                <Toast show={toast.show} onClose={() => {
                    let tmp = toastQueue.slice();
                    tmp[i].show = false;
                    setToastQueue(tmp);
                }} >
                    <Toast.Header>
                        <strong className="me-auto">{toast.title}</strong>
                    </Toast.Header>
                    <Toast.Body>{toast.text}</Toast.Body>
                </Toast>
                )
            })
        }


        
    }

    return (
        <div>
            <Navbar bg="dark" variant="dark" expand={true}>

                <LinkContainer to="/">
                    <Navbar.Brand>
                        Admin Panel
                    </Navbar.Brand>
                </LinkContainer>
                <Navbar.Toggle />
                <Navbar.Collapse>
                    <Nav>
                        <LinkContainer to="home">
                            <Nav.Link>
                                Home
                            </Nav.Link>
                        </LinkContainer>
                        {currentUser && (
                            <LinkContainer to="/user">
                                <Nav.Link>
                                    Users
                                </Nav.Link>
                            </LinkContainer>
                        )}
                    </Nav>

                    {currentUser ? (
                        <Nav className="justify-content-end" style={{ width: "100%" }}>
                            <Nav.Item>
                                <LinkContainer to="/profile">
                                    <Nav.Link>
                                        {currentUser.user.name}
                                    </Nav.Link>
                                </LinkContainer>
                            </Nav.Item>

                            <Nav.Item>
                                <LinkContainer to="/login" onClick={logOut}>
                                    <Nav.Link>
                                        log out
                                    </Nav.Link>
                                </LinkContainer>
                            </Nav.Item>
                        </Nav>
                    ) : (
                        <div className="navbar-nav ml-auto">
                            <li className="nav-item">
                                <Link to={"/login"} className="nav-link">
                                    Login
                                </Link>
                            </li>
                        </div>
                    )}
                </Navbar.Collapse>
            </Navbar>

            <div className="container mt-3">
                <Switch>
                    <Route exact path={["/", "/home"]} component={Home} />
                    <Route exact path="/login" component={Login} />

                    <Route path="/user" render={(props) => <BoardUser {...props} addToast={addToast} /> } />

                </Switch>

            </div>
            <ToastContainer className="p-3" position="bottom-center">
                {toastQueue && renderToasts()}
            </ToastContainer>
        </div>
    );
};

export default App;
