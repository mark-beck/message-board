import { useState } from "react";
import { Modal } from "react-bootstrap";
import { Button } from "react-bootstrap";
import { Form } from "react-bootstrap";
import userService from "../../services/user.service";

const CreateForm = (props) => {

    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [user, setUser] = useState(true);
    const [moderator, setModerator] = useState(false);
    const [admin, setAdmin] = useState(false);

    const handleCreate = () => {
        console.log("name: " + name);
        console.log("email: " + email);
        console.log("password: " + password);
        console.log("user: " + user);
        console.log("moderator: " + moderator);
        console.log("admin: " + admin);
        userService.createUser(name, email, password, user, moderator, admin).then(() => props.reload_list());
        handleClose();
    }

    const handleClose = () => {
        setName("");
        setEmail("");
        setUser(true);
        setModerator(false);
        setAdmin(false);
        props.handleClose();
    }

    return (
        <Modal show={props.show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>create new User</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group className="mb-3" controlId="createForm.name">
                        <Form.Label>Name</Form.Label>
                        <Form.Control
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="createForm.email">
                        <Form.Label>email</Form.Label>
                        <Form.Control
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="createForm.password">
                        <Form.Label>email</Form.Label>
                        <Form.Control
                            type="email"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="createForm.boxes">
                        <Form.Label>Roles</Form.Label>
                        <Form.Check
                            type="checkbox"
                            id="createForm.boxes.user"
                            label="User"
                            checked={user}
                            onChange={(e) => setUser(e.target.checked)}
                        />
                        <Form.Check
                            type="checkbox"
                            id="createForm.boxes.moderator"
                            label="Moderator"
                            checked={moderator}
                            onChange={(e) => setModerator(e.target.checked)}
                        />
                        <Form.Check
                            type="checkbox"
                            id="createForm.boxes.admin"
                            label="Admin"
                            checked={admin}
                            onChange={(e) => setAdmin(e.target.checked)}
                        />
                    </Form.Group>
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={handleClose}>
                    Close
                </Button>
                <Button variant="primary" onClick={handleCreate}>
                    Create
                </Button>
            </Modal.Footer>
        </Modal>
    );
}

export default CreateForm;