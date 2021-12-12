import { useState } from "react";
import { Modal } from "react-bootstrap";
import { Button } from "react-bootstrap";
import { Form } from "react-bootstrap";
import userService from "../services/user.service";

const CreateForm = (props) => {

    const [author, setAuthor] = useState("");
    const [text, setText] = useState("");
    const [date, setDate] = useState("");
    

    const handleCreate = () => {
        
        console.log("author: " + author);
        console.log("text: " + text);
        userService.postContent(author, text, date).then(() => props.reload_posts());
        handleClose();
    }

    const handleClose = () => {
        props.handleClose();
    }

    return (
        <Modal show={props.show} onHide={handleClose}>
            <Modal.Header closeButton>
                <Modal.Title>write a new Post</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group className="mb-3" controlId="createForm.author">
                        <Form.Label>author</Form.Label>
                        <Form.Control
                            type="text"
                            value={author}
                            onChange={(e) => setAuthor(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="createForm.text">
                        <Form.Label>text</Form.Label>
                        <Form.Control
                            type="text"
                            value={text}
                            onChange={(e) => setText(e.target.value)}
                        />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="createForm.date">
                        <Form.Label>date</Form.Label>
                        <Form.Control
                            type="text"
                            value={date}
                            onChange={(e) => setDate(e.target.value)}
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