import json
from flask_wtf import FlaskForm
from wtforms import StringField, SubmitField, TextAreaField, BooleanField
from wtforms.validators import DataRequired, ValidationError


class UserConfig(FlaskForm):
    auth_spec = StringField('Auth Spec', validators=[DataRequired()])
    user_config = TextAreaField('User Config', validators=[DataRequired()])
    sample_report = BooleanField('Sample Report')
    submit = SubmitField('Submit')

    def validate_user_config(form, field):
        user_config = json.loads(field.data)
        expected_keys = {'locale', 'injection_targets', 'observers'}
        if not expected_keys.issubset(set(user_config)):
            message = f'missing fields in config object {expected_keys - set(user_config)}'
            raise ValidationError(message)
