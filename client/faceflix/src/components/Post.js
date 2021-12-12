import { useEffect, useState } from "react"
import { MDBCard, MDBCardBody, MDBRow, MDBCardTitle, MDBCardText, MDBBtn, MDBIcon } from 'mdb-react-ui-kit';


const Post = (props) => {
    const [author, setAuthor] = useState(props.author)
    const [text, setText] = useState(props.text)

    useEffect(() => {
        setAuthor(props.author)
        setText(props.text)
    }, [props.author, props.text]);

    return (
        <MDBCard
            className="my-5 px-5 pt-4"
            style={{ fontWeight: 300, maxWidth: 600 }}
        >
            <MDBCardBody className="py-0">
                <MDBRow>
                    <div className="mdb-feed">
                        <div className="news">
                            <div className="label">
                                {/* <img
                    src="//ssl.gstatic.com/accounts/ui/avatar_1x.png"
                    alt=""
                    className="rounded-circle z-depth-1-half"
                /> */}
                            </div>
                            <div className="excerpt">
                                <div className="brief">
                                    <a href="#!" className="name">
                                        {author}
                                    </a>
                                </div>
                                <div className="added-text">
                                    {text}
                                </div>
                            </div>
                        </div>
                    </div>
                </MDBRow>
            </MDBCardBody>
        </MDBCard>

    );

}

export default Post;