import { useEffect, useState } from "react"
import { MDBCard, MDBCardBody, MDBCardTitle, MDBCardText, MDBBtn } from 'mdb-react-ui-kit';


const Post = (props) => {
    const [author, setAuthor] = useState(props.author)
    const [text, setText] = useState(props.text)

    return (
        <MDBCard style={{ maxWidth: '22rem' }}>
            <MDBCardBody>
                <MDBCardTitle>{author}</MDBCardTitle>
                <MDBCardText>
                    {text}
                </MDBCardText>
            </MDBCardBody>
        </MDBCard>
    );

}

export default Post;