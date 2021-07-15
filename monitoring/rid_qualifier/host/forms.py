from flask_wtf import FlaskForm
from wtforms import StringField, PasswordField, BooleanField, SubmitField, TextField
from wtforms.validators import DataRequired

class UserConfig(FlaskForm):
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextField('User Config', validators=[DataRequired()])
    submit = SubmitField('Submit')

    # username = StringField('Username', validators=[DataRequired()])
    # password = PasswordField('Password', validators=[DataRequired()])
    # remember_me = BooleanField('Remember Me')
    # submit = SubmitField('Sign In')