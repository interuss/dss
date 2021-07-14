from flask_wtf import FlaskForm
from wtforms import StringField, PasswordField, BooleanField, SubmitField
from wtforms.validators import DataRequired

class UserConfig(FlaskForm):
    username = StringField('auth_spec', validators=[DataRequired()])
    password = PasswordField('user_config', validators=[DataRequired()])
    submit = SubmitField('Submit')