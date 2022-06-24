import { useEffect, useState } from "react"
import { MDBCard, MDBCardBody, MDBRow, MDBCardTitle, MDBCardText, MDBBtn, MDBIcon, MDBCol } from 'mdb-react-ui-kit';
import { formatDate, delete_post } from "../services/user.service";
import { getCurrentUser } from "../services/auth.service";
import { Button } from "react-bootstrap";


const Post = (props) => {
    const [postData, setPostData] = useState(props.postData)
    const [owned, setOwned] = useState(postData.author === getCurrentUser().user.name)


    useEffect(() => {
        setPostData(props.postData)
        setOwned(postData.author === getCurrentUser().user.name)
    }, [props.postData]);

    return (
        <MDBCard
            style={{ fontWeight: 300, maxWidth: 600 }}
        >
            <MDBCardBody>
                <p>
                    {postData.text}
                </p>
                <div className="d-flex justify-content-between">
                    <div className="d-flex flex-row align-items-center">
                        <p className="small mb-0 ms-2">{postData.author}</p>
                    </div>
                    <div className="d-flex flex-row align-items-center">
                        <p className="small text-muted mb-0">{formatDate(postData.date)}</p>
                        <i className="far fa-thumbs-up ms-2 fa-xs text-black" style={{ marginTop: -0.16 + 'rem' }}></i>
                    </div>
                </div>
                {owned &&
                    <MDBBtn variant="primary" onClick={() => {
                        delete_post(postData.id).then(() => (
                            window.location.reload()
                        )).catch((error) => {
                            console.log(error)
                        })
                    }}>
                        Delete
                    </MDBBtn>
                }
            </MDBCardBody >
        </MDBCard >

    );

}

export default Post;