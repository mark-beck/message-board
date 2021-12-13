import { Table, Button } from "react-bootstrap";
import userService from "../../services/user.service";
import { useState, useEffect } from "react";
import { Trash } from "react-bootstrap-icons";
import App from "../../App";

const UserList = (props) => {
    console.log(typeof (props.userslist));
    
    const [userList, setUserList] = useState(props.userlist);

    useEffect(() => {
        setUserList(props.userlist);
    }, [props.userlist]);

    const handleDelete = (name) => {
        userService.deleteUser(name).then(() => {
            props.reload_list()
        });
    }
    
    const inner = userList.map((data, i) => {
        return (
            <tr>
                <th scope="row">{i}</th>
                <td>{data.name}</td>
                <td>{data.email}</td>
                <td>{data.roles.join(",")}</td>
                <td>
                    <Button onClick={() => handleDelete(data.name)}>
                        delete
                    </Button>
                </td>
            </tr>
        );
    });
    return (
        <Table hover>
            <thead>
                <tr>
                    <th scope="col">#</th>
                    <th scope="col">Name</th>
                    <th scope="col">Email</th>
                    <th scope="col">Roles</th>
                    <th scope="col">Delete</th>
                </tr>
            </thead>
            <tbody>
                {inner}
            </tbody>
        </Table>
    )

}

export default UserList;