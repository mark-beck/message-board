import React, { useState, useEffect } from "react";

import { Modal } from "react-bootstrap";
import { Button } from "react-bootstrap";
import { Form, Col, Container, Row } from "react-bootstrap";
import Post from "./Post";
import userService from "../services/user.service";

const Search = () => {
    const [searchtext, setSearchtext] = useState("");
    const [author, setAuthor] = useState("");
    const [sortBy, setSortBy] = useState("");
    const [sortDir, setSortDir] = useState("");
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const search = () => {
        setLoading(true);
        userService.getFilteredContent({ text: searchtext, author: author, sortBy: sortBy, sortOrder: sortDir }).then((results) => {
            setResults(results);
            setLoading(false);
        }).catch((error) => {
            setError(error);
            setLoading(false);
        }
        );
    }

    return (
        <>
            <div>
                <Container>
                    <Form>
                        <Form.Group className="mb-3" controlId="filterform.text">
                            <Form.Label>searchtext</Form.Label>
                            <Form.Control
                                type="text"
                                value={searchtext}
                                onChange={(e) => setSearchtext(e.target.value)}
                            />
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="filterform.author">
                            <Form.Label>Author</Form.Label>
                            <Form.Control
                                type="text"
                                value={author}
                                onChange={(e) => setAuthor(e.target.value)}
                            />
                        </Form.Group>
                        <Row>
                            <Form.Group as={Col} controlId="filterform.sortBy">
                                <Form.Label>Sort By</Form.Label>
                                <Form.Select

                                    defaultValue="date"
                                    onChange={(e) => setSortBy(e.target.value)}
                                >
                                    <option value="date">date</option>
                                    <option value="author">author</option>
                                    <option value="text">text</option>
                                </Form.Select>
                            </Form.Group>
                            <Form.Group as={Col} controlId="filterform.sortDir">
                                <Form.Label>Sort Dir</Form.Label>
                                <Form.Select
                                    value={sortDir}
                                    onChange={(e) => setSortDir(e.target.value)}
                                >
                                    <option value="asc">asc</option>
                                    <option value="desc">desc</option>
                                </Form.Select>
                            </Form.Group>

                        </Row>
                        <Form.Group className="mb-3" controlId="filterform.submit">
                            <Button variant="primary" onClick={() => search()}>
                                search
                            </Button>
                        </Form.Group>
                    </Form>

                </Container>
            </div>
            <div>
                {loading && <div>Loading...</div>}
                {error && <div>Error: {error.message}</div>}
                {results.map((post) => {
                    return (
                        <Post postData={post} />
                    );
                })}

            </div>
        </>
    );

}

export default Search;

